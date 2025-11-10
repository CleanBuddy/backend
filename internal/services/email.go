package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// EmailService handles sending emails via Sidemail API
type EmailService struct {
	apiKey     string
	apiURL     string
	fromName   string
	fromAddress string
	httpClient *http.Client
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	// Support both EMAIL_API_KEY (new) and SIDEMAIL_API_KEY (legacy)
	apiKey := os.Getenv("EMAIL_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SIDEMAIL_API_KEY") // Legacy fallback
	}

	return &EmailService{
		apiKey:      apiKey,
		apiURL:      "https://api.sidemail.io/v1/email/send",
		fromName:    os.Getenv("EMAIL_FROM_NAME"),
		fromAddress: os.Getenv("EMAIL_FROM_ADDRESS"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// EmailRequest represents the request to Sidemail API
type EmailRequest struct {
	ToAddress     string                 `json:"toAddress"`
	FromAddress   string                 `json:"fromAddress"`
	FromName      string                 `json:"fromName"`
	TemplateName  string                 `json:"templateName,omitempty"`
	TemplateProps map[string]interface{} `json:"templateProps,omitempty"`
	Subject       string                 `json:"subject,omitempty"`
	HTML          string                 `json:"html,omitempty"`
	Text          string                 `json:"text,omitempty"`
}

// EmailResponse represents the response from Sidemail API
type EmailResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// SendEmail sends an email via Sidemail API
func (s *EmailService) SendEmail(ctx context.Context, req EmailRequest) (*EmailResponse, error) {
	// Set default from address and name if not provided
	if req.FromAddress == "" {
		req.FromAddress = s.fromAddress
	}
	if req.FromName == "" {
		req.FromName = s.fromName
	}

	// Validate required fields
	if req.ToAddress == "" {
		return nil, fmt.Errorf("toAddress is required")
	}
	if req.FromAddress == "" {
		return nil, fmt.Errorf("fromAddress is required")
	}

	// In development mode without API key, just log the email
	if os.Getenv("ENV") == "development" && s.apiKey == "" {
		fmt.Printf("\n📧 === EMAIL (Development Mode) ===\n")
		fmt.Printf("To: %s\n", req.ToAddress)
		fmt.Printf("From: %s <%s>\n", req.FromName, req.FromAddress)
		fmt.Printf("Subject: %s\n", req.Subject)
		if req.HTML != "" {
			fmt.Printf("HTML Body: %s\n", req.HTML)
		}
		if req.Text != "" {
			fmt.Printf("Text Body: %s\n", req.Text)
		}
		if req.TemplateName != "" {
			fmt.Printf("Template: %s\n", req.TemplateName)
			fmt.Printf("Template Props: %+v\n", req.TemplateProps)
		}
		fmt.Printf("=====================================\n\n")
		return &EmailResponse{ID: "dev-mode-" + time.Now().Format("20060102150405"), Status: "development"}, nil
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("email API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var emailResp EmailResponse
	if err := json.Unmarshal(body, &emailResp); err != nil {
		return nil, fmt.Errorf("failed to parse email response: %w", err)
	}

	return &emailResp, nil
}

// SendBookingConfirmationEmail sends a booking confirmation email to the client
func (s *EmailService) SendBookingConfirmationEmail(ctx context.Context, toEmail, clientName, bookingID, serviceType, scheduledDate, scheduledTime string, totalPrice float64) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "booking-confirmation",
		TemplateProps: map[string]interface{}{
			"clientName":    clientName,
			"bookingID":     bookingID,
			"serviceType":   serviceType,
			"scheduledDate": scheduledDate,
			"scheduledTime": scheduledTime,
			"totalPrice":    fmt.Sprintf("%.2f RON", totalPrice),
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendBookingAcceptedEmail sends email when cleaner accepts a booking
func (s *EmailService) SendBookingAcceptedEmail(ctx context.Context, toEmail, clientName, cleanerName, bookingID, scheduledDate, scheduledTime string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "booking-accepted",
		TemplateProps: map[string]interface{}{
			"clientName":    clientName,
			"cleanerName":   cleanerName,
			"bookingID":     bookingID,
			"scheduledDate": scheduledDate,
			"scheduledTime": scheduledTime,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendBookingCancelledEmail sends email when booking is cancelled
func (s *EmailService) SendBookingCancelledEmail(ctx context.Context, toEmail, recipientName, bookingID, cancelledBy, cancellationReason string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "booking-cancelled",
		TemplateProps: map[string]interface{}{
			"recipientName":      recipientName,
			"bookingID":          bookingID,
			"cancelledBy":        cancelledBy,
			"cancellationReason": cancellationReason,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendBookingCompletedEmail sends email when booking is completed
func (s *EmailService) SendBookingCompletedEmail(ctx context.Context, toEmail, clientName, bookingID string, totalPrice float64, reviewURL string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "booking-completed",
		TemplateProps: map[string]interface{}{
			"clientName": clientName,
			"bookingID":  bookingID,
			"totalPrice": fmt.Sprintf("%.2f RON", totalPrice),
			"reviewURL":  reviewURL,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendCleanerApprovedEmail sends email when cleaner is approved
func (s *EmailService) SendCleanerApprovedEmail(ctx context.Context, toEmail, cleanerName string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "cleaner-approved",
		TemplateProps: map[string]interface{}{
			"cleanerName": cleanerName,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendCleanerRejectedEmail sends email when cleaner application is rejected
func (s *EmailService) SendCleanerRejectedEmail(ctx context.Context, toEmail, cleanerName, rejectionReason string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "cleaner-rejected",
		TemplateProps: map[string]interface{}{
			"cleanerName":     cleanerName,
			"rejectionReason": rejectionReason,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendPayoutProcessedEmail sends email when payout is processed
func (s *EmailService) SendPayoutProcessedEmail(ctx context.Context, toEmail, cleanerName string, amount float64, period, transferRef string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "payout-processed",
		TemplateProps: map[string]interface{}{
			"cleanerName": cleanerName,
			"amount":      fmt.Sprintf("%.2f RON", amount),
			"period":      period,
			"transferRef": transferRef,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendWelcomeEmail sends welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail, userName, userRole string) error {
	req := EmailRequest{
		ToAddress:    toEmail,
		TemplateName: "welcome",
		TemplateProps: map[string]interface{}{
			"userName": userName,
			"userRole": userRole,
		},
	}

	_, err := s.SendEmail(ctx, req)
	return err
}

// SendOTPEmail sends an OTP verification code email
func (s *EmailService) SendOTPEmail(ctx context.Context, toEmail, code string) error {
	// HTML email body with Romanian content
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ro">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f5f5f5;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; border-radius: 10px 10px 0 0; text-align: center;">
        <h1 style="color: white; margin: 0; font-size: 28px;">CleanBuddy</h1>
        <p style="color: rgba(255,255,255,0.9); margin: 10px 0 0 0;">Servicii de curățenie la cerere</p>
    </div>

    <div style="background: white; padding: 40px; border-radius: 0 0 10px 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #667eea; margin-top: 0;">Codul tău de verificare</h2>

        <p>Bună!</p>

        <p>Ai solicitat un cod de verificare pentru a te autentifica în CleanBuddy. Iată codul tău:</p>

        <div style="background: #f7f9fc; border: 2px solid #667eea; border-radius: 8px; padding: 20px; text-align: center; margin: 30px 0;">
            <span style="font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px; font-family: 'Courier New', monospace;">%s</span>
        </div>

        <p><strong>Acest cod este valabil 5 minute.</strong></p>

        <p>Dacă nu ai solicitat acest cod, poți ignora acest email în siguranță.</p>

        <hr style="border: none; border-top: 1px solid #e0e0e0; margin: 30px 0;">

        <p style="color: #888; font-size: 14px; margin: 0;">
            Cu drag,<br>
            Echipa CleanBuddy
        </p>
    </div>

    <div style="text-align: center; margin-top: 20px; color: #888; font-size: 12px;">
        <p>CleanBuddy - Servicii de curățenie profesionale în România</p>
        <p>© 2025 CleanBuddy. Toate drepturile rezervate.</p>
    </div>
</body>
</html>
`, code)

	// Plain text version
	textBody := fmt.Sprintf(`
CleanBuddy - Codul tău de verificare

Bună!

Ai solicitat un cod de verificare pentru a te autentifica în CleanBuddy.

Codul tău este: %s

Acest cod este valabil 5 minute.

Dacă nu ai solicitat acest cod, poți ignora acest email în siguranță.

Cu drag,
Echipa CleanBuddy

---
CleanBuddy - Servicii de curățenie profesionale în România
© 2025 CleanBuddy. Toate drepturile rezervate.
`, code)

	req := EmailRequest{
		ToAddress: toEmail,
		Subject:   "Codul tău de verificare CleanBuddy",
		HTML:      htmlBody,
		Text:      textBody,
	}

	_, err := s.SendEmail(ctx, req)
	return err
}
