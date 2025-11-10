package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
)

// ANAFClient handles communication with ANAF e-Factura SPV (Spatiu Privat Virtual)
type ANAFClient struct {
	config      *config.ANAFConfig
	companyConfig *config.CompanyConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// ANAFUploadResponse represents the response from ANAF after uploading an invoice
type ANAFUploadResponse struct {
	UploadIndex string      `json:"upload_index"` // Unique identifier for this upload
	Status      string      `json:"status"`       // "processing", "accepted", "rejected"
	Message     string      `json:"message,omitempty"`
	Errors      []ANAFError `json:"errors,omitempty"`
}

// ANAFStatusResponse represents the status of an uploaded invoice
type ANAFStatusResponse struct {
	UploadIndex   string    `json:"upload_index"`
	Status        string    `json:"status"` // "processing", "accepted", "rejected"
	DateUploaded  time.Time `json:"date_uploaded"`
	DateProcessed time.Time `json:"date_processed,omitempty"`
	DownloadID    string    `json:"download_id,omitempty"` // ID for confirmation PDF
	Errors        []ANAFError `json:"errors,omitempty"`
}

// ANAFError represents an error from ANAF
type ANAFError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// tokenResponse represents OAuth2 token response
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // Seconds until expiration
	Scope       string `json:"scope"`
}

// NewANAFClient creates a new ANAF API client
func NewANAFClient(anafConfig *config.ANAFConfig, companyConfig *config.CompanyConfig) *ANAFClient {
	return &ANAFClient{
		config:        anafConfig,
		companyConfig: companyConfig,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getBaseURL returns the appropriate ANAF API base URL (sandbox or production)
func (c *ANAFClient) getBaseURL() string {
	if c.config.Environment == "production" {
		return c.config.ProductionURL
	}
	return c.config.SandboxURL
}

// getAccessToken retrieves a valid OAuth2 access token (cached or new)
func (c *ANAFClient) getAccessToken(ctx context.Context) (string, error) {
	// Check if we have a valid cached token
	c.tokenMutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		token := c.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// Need to acquire new token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		return c.accessToken, nil
	}

	// Request new OAuth2 token using client credentials flow
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("scope", "efactura")

	tokenURL := c.getBaseURL() + "/oauth/token"
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Cache the token
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return c.accessToken, nil
}

// UploadInvoice uploads a UBL XML invoice to ANAF SPV
func (c *ANAFClient) UploadInvoice(ctx context.Context, xmlContent []byte, invoiceNumber string) (*ANAFUploadResponse, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add XML file
	part, err := writer.CreateFormFile("file", fmt.Sprintf("%s.xml", invoiceNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(xmlContent); err != nil {
		return nil, fmt.Errorf("failed to write XML content: %w", err)
	}

	// Add CUI (company tax ID)
	if err := writer.WriteField("cif", c.companyConfig.CUI); err != nil {
		return nil, fmt.Errorf("failed to write CIF field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create upload request
	uploadURL := c.getBaseURL() + "/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload invoice: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var uploadResp ANAFUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &uploadResp, nil
}

// GetInvoiceStatus checks the status of a previously uploaded invoice
func (c *ANAFClient) GetInvoiceStatus(ctx context.Context, uploadIndex string) (*ANAFStatusResponse, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create status request
	statusURL := fmt.Sprintf("%s/status/%s", c.getBaseURL(), uploadIndex)
	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp ANAFStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &statusResp, nil
}

// DownloadConfirmation downloads the ANAF confirmation PDF for an accepted invoice
func (c *ANAFClient) DownloadConfirmation(ctx context.Context, downloadID string) ([]byte, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create download request
	downloadURL := fmt.Sprintf("%s/download/%s", c.getBaseURL(), downloadID)
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download confirmation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read PDF content
	pdfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF content: %w", err)
	}

	return pdfContent, nil
}

// RetryUpload retries uploading an invoice with exponential backoff
func (c *ANAFClient) RetryUpload(ctx context.Context, xmlContent []byte, invoiceNumber string, maxRetries int) (*ANAFUploadResponse, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds
			backoffDuration := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		resp, err := c.UploadInvoice(ctx, xmlContent, invoiceNumber)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		// Only retry on network errors or 5xx server errors
		// Don't retry on validation errors (4xx)
	}

	return nil, fmt.Errorf("upload failed after %d attempts: %w", maxRetries, lastErr)
}
