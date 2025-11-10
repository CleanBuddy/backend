package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/models"
)

// InvoiceService handles invoice business logic
type InvoiceService struct {
	invoiceRepo  *models.InvoiceRepository
	bookingRepo  *models.BookingRepository
	userRepo     *models.UserRepository
	pdfGenerator *PDFGenerator
	xmlGenerator *XMLGenerator
	anafClient   *ANAFClient
	config       *config.CompanyConfig
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *sql.DB, companyConfig *config.CompanyConfig, anafConfig *config.ANAFConfig) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:  models.NewInvoiceRepository(db),
		bookingRepo:  models.NewBookingRepository(db),
		userRepo:     models.NewUserRepository(db),
		pdfGenerator: NewPDFGenerator("./invoices/pdf"),
		xmlGenerator: NewXMLGenerator("./invoices/xml", companyConfig),
		anafClient:   NewANAFClient(anafConfig, companyConfig),
		config:       companyConfig,
	}
}

// CreateInvoiceForBooking auto-creates invoice when booking is completed
func (s *InvoiceService) CreateInvoiceForBooking(bookingID string) (*models.Invoice, error) {
	// Get booking details
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if invoice already exists
	existing, err := s.invoiceRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invoice: %w", err)
	}
	if existing != nil {
		return existing, nil // Invoice already exists
	}

	// Get client details
	client, err := s.userRepo.GetByID(booking.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	// Get cleaner details
	var cleanerName string
	if booking.CleanerID.Valid {
		cleaner, err := s.userRepo.GetByID(booking.CleanerID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to get cleaner: %w", err)
		}
		if cleaner != nil {
			cleanerName = fmt.Sprintf("%s %s",
				cleaner.FirstName.String,
				cleaner.LastName.String)
		}
	}
	if cleanerName == "" {
		cleanerName = "CleanBuddy Cleaner"
	}

	// Generate invoice number
	invoiceNumber, err := s.GenerateInvoiceNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Prepare client name
	clientName := fmt.Sprintf("%s %s",
		client.FirstName.String,
		client.LastName.String)
	if clientName == " " {
		clientName = "Client"
	}

	// Prepare service description
	serviceDescription := s.buildServiceDescription(booking)

	// Calculate dates
	issueDate := time.Now()
	dueDate := issueDate.AddDate(0, 0, 14) // 14 days payment term

	// Create invoice
	invoice := &models.Invoice{
		BookingID:          bookingID,
		InvoiceNumber:      invoiceNumber,
		IssueDate:          issueDate,
		DueDate:            dueDate,
		ClientName:         clientName,
		CleanerName:        cleanerName,
		ServiceDescription: serviceDescription,
		Subtotal:           booking.TotalPrice,
		TaxAmount:          0, // No tax for now
		TotalAmount:        booking.TotalPrice,
		Currency:           "RON",
		Status:             models.InvoiceStatusIssued,
	}

	// Add client email if available
	if client.Email.Valid {
		invoice.ClientEmail = sql.NullString{String: client.Email.String, Valid: true}
	}

	// Create invoice in database
	if err := s.invoiceRepo.Create(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Generate PDF
	pdfPath, err := s.pdfGenerator.GenerateInvoicePDF(invoice)
	if err != nil {
		// Log error but don't fail invoice creation
		fmt.Printf("Warning: failed to generate PDF for invoice %s: %v\n", invoice.ID, err)
	} else {
		invoice.PdfURL = sql.NullString{String: pdfPath, Valid: true}
	}

	// Generate XML for ANAF e-Factura
	xmlPath, err := s.xmlGenerator.GenerateInvoiceXML(invoice, booking)
	if err != nil {
		// Log error but don't fail invoice creation
		fmt.Printf("Warning: failed to generate XML for invoice %s: %v\n", invoice.ID, err)
	} else {
		invoice.XmlURL = sql.NullString{String: xmlPath, Valid: true}
	}

	// Update invoice with file paths
	if invoice.PdfURL.Valid || invoice.XmlURL.Valid {
		if err := s.invoiceRepo.Update(invoice); err != nil {
			fmt.Printf("Warning: failed to update invoice %s with file paths: %v\n", invoice.ID, err)
		}
	}

	// Auto-submit to ANAF if enabled in config
	// Note: This runs asynchronously to avoid blocking invoice creation
	go func() {
		if err := s.SubmitToANAF(invoice.ID); err != nil {
			fmt.Printf("Warning: failed to auto-submit invoice %s to ANAF: %v\n", invoice.InvoiceNumber, err)
			// Error is logged but doesn't block invoice creation
			// Invoice will be picked up by the batch processor later
		} else {
			fmt.Printf("Successfully auto-submitted invoice %s to ANAF\n", invoice.InvoiceNumber)
		}
	}()

	return invoice, nil
}

// GenerateInvoiceNumber generates next invoice number
func (s *InvoiceService) GenerateInvoiceNumber() (string, error) {
	return s.invoiceRepo.GenerateInvoiceNumber()
}

// GetInvoiceByID gets an invoice by ID (no auth check - for internal use)
func (s *InvoiceService) GetInvoiceByID(invoiceID string) (*models.Invoice, error) {
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}
	return invoice, nil
}

// GetInvoice gets an invoice by ID (with auth check)
func (s *InvoiceService) GetInvoice(invoiceID string, userID string) (*models.Invoice, error) {
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Verify ownership through booking
	booking, err := s.bookingRepo.GetByID(invoice.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if user is client or cleaner
	if booking.ClientID != userID {
		if !booking.CleanerID.Valid || booking.CleanerID.String != userID {
			return nil, fmt.Errorf("unauthorized")
		}
	}

	return invoice, nil
}

// GetInvoiceByBookingID gets an invoice by booking ID
func (s *InvoiceService) GetInvoiceByBookingID(bookingID string, userID string) (*models.Invoice, error) {
	// Verify ownership through booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if user is client or cleaner
	if booking.ClientID != userID {
		if !booking.CleanerID.Valid || booking.CleanerID.String != userID {
			return nil, fmt.Errorf("unauthorized")
		}
	}

	invoice, err := s.invoiceRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetMyInvoices gets all invoices for a user
func (s *InvoiceService) GetMyInvoices(userID string) ([]*models.Invoice, error) {
	return s.invoiceRepo.GetByUserID(userID)
}

// buildServiceDescription builds a human-readable service description
func (s *InvoiceService) buildServiceDescription(booking *models.Booking) string {
	desc := fmt.Sprintf("Serviciu de curățenie - %s", s.translateServiceType(booking.ServiceType))

	if booking.AreaSqm.Valid {
		desc += fmt.Sprintf(", Suprafață: %d mp", booking.AreaSqm.Int32)
	}

	desc += fmt.Sprintf(", Durata estimată: %d ore", booking.EstimatedHours)

	// Add add-ons
	addons := []string{}
	if booking.IncludesDeepCleaning {
		addons = append(addons, "curățenie profundă")
	}
	if booking.IncludesWindows {
		addons = append(addons, fmt.Sprintf("geamuri (%d)", booking.NumberOfWindows))
	}
	if booking.IncludesCarpetCleaning {
		addons = append(addons, fmt.Sprintf("covoare (%d mp)", booking.CarpetAreaSqm))
	}

	if len(addons) > 0 {
		desc += ", Include: "
		for i, addon := range addons {
			if i > 0 {
				desc += ", "
			}
			desc += addon
		}
	}

	// Add date
	desc += fmt.Sprintf(", Data: %s", booking.ScheduledDate.Format("02.01.2006"))

	return desc
}

// translateServiceType translates service type to Romanian
func (s *InvoiceService) translateServiceType(serviceType models.ServiceType) string {
	switch serviceType {
	case models.ServiceTypeStandard:
		return "Standard"
	case models.ServiceTypeDeepCleaning:
		return "Curățenie Profundă"
	case models.ServiceTypeOffice:
		return "Birou"
	case models.ServiceTypePostRenovation:
		return "Post-Renovare"
	case models.ServiceTypeMoveInOut:
		return "Mutare"
	default:
		return string(serviceType)
	}
}

// SubmitToANAF submits an invoice to ANAF e-Factura system
func (s *InvoiceService) SubmitToANAF(invoiceID string) error {
	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Check if already submitted
	if invoice.ANAFStatus == models.ANAFStatusAccepted || invoice.ANAFStatus == models.ANAFStatusProcessing {
		return fmt.Errorf("invoice already submitted to ANAF")
	}

	// Get booking details for XML generation
	booking, err := s.bookingRepo.GetByID(invoice.BookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	// Read or generate XML content
	xmlPath := fmt.Sprintf("./invoices/xml/%s.xml", invoice.InvoiceNumber)
	xmlContent, err := s.xmlGenerator.ReadXML(xmlPath)
	if err != nil {
		// Try to regenerate XML if not found
		_, genErr := s.xmlGenerator.GenerateInvoiceXML(invoice, booking)
		if genErr != nil {
			return fmt.Errorf("failed to generate XML: %w", genErr)
		}
		xmlContent, err = s.xmlGenerator.ReadXML(xmlPath)
		if err != nil {
			return fmt.Errorf("failed to read XML: %w", err)
		}
	}

	// Submit to ANAF
	ctx := context.Background()
	resp, err := s.anafClient.UploadInvoice(ctx, xmlContent, invoice.InvoiceNumber)
	if err != nil {
		// Update status to failed
		invoice.ANAFStatus = models.ANAFStatusFailed
		invoice.ANAFRetryCount++
		invoice.ANAFLastRetryAt = sql.NullTime{Time: time.Now(), Valid: true}
		if updateErr := s.invoiceRepo.UpdateANAFStatus(invoice); updateErr != nil {
			return fmt.Errorf("failed to update ANAF status after error: %w", updateErr)
		}
		return fmt.Errorf("failed to upload invoice to ANAF: %w", err)
	}

	// Update invoice with ANAF response
	invoice.ANAFUploadIndex = sql.NullString{String: resp.UploadIndex, Valid: true}
	invoice.ANAFStatus = models.ANAFStatus(resp.Status)
	invoice.ANAFSubmittedAt = sql.NullTime{Time: time.Now(), Valid: true}

	// Store errors if any
	if len(resp.Errors) > 0 {
		invoice.ANAFErrors = make([]models.ANAFError, len(resp.Errors))
		for i, err := range resp.Errors {
			invoice.ANAFErrors[i] = models.ANAFError{
				Code:    err.Code,
				Message: err.Message,
				Field:   err.Field,
			}
		}
		invoice.ANAFStatus = models.ANAFStatusRejected
	} else if resp.Status == "processing" {
		invoice.ANAFStatus = models.ANAFStatusProcessing
	} else if resp.Status == "accepted" {
		invoice.ANAFStatus = models.ANAFStatusAccepted
		invoice.ANAFProcessedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	// Save to database
	if err := s.invoiceRepo.UpdateANAFStatus(invoice); err != nil {
		return fmt.Errorf("failed to update invoice ANAF status: %w", err)
	}

	return nil
}

// CheckANAFStatus checks the status of an invoice in ANAF system
func (s *InvoiceService) CheckANAFStatus(invoiceID string) error {
	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Check if invoice has been submitted
	if !invoice.ANAFUploadIndex.Valid || invoice.ANAFUploadIndex.String == "" {
		return fmt.Errorf("invoice not submitted to ANAF")
	}

	// Query ANAF for status
	ctx := context.Background()
	statusResp, err := s.anafClient.GetInvoiceStatus(ctx, invoice.ANAFUploadIndex.String)
	if err != nil {
		return fmt.Errorf("failed to check ANAF status: %w", err)
	}

	// Update invoice status
	invoice.ANAFStatus = models.ANAFStatus(statusResp.Status)
	if statusResp.Status == "accepted" {
		invoice.ANAFStatus = models.ANAFStatusAccepted
		invoice.ANAFProcessedAt = sql.NullTime{Time: statusResp.DateProcessed, Valid: true}
		if statusResp.DownloadID != "" {
			invoice.ANAFDownloadID = sql.NullString{String: statusResp.DownloadID, Valid: true}
		}
	} else if statusResp.Status == "rejected" {
		invoice.ANAFStatus = models.ANAFStatusRejected
		invoice.ANAFProcessedAt = sql.NullTime{Time: statusResp.DateProcessed, Valid: true}
		// Store errors
		if len(statusResp.Errors) > 0 {
			invoice.ANAFErrors = make([]models.ANAFError, len(statusResp.Errors))
			for i, err := range statusResp.Errors {
				invoice.ANAFErrors[i] = models.ANAFError{
					Code:    err.Code,
					Message: err.Message,
					Field:   err.Field,
				}
			}
		}
	}

	// Save to database
	if err := s.invoiceRepo.UpdateANAFStatus(invoice); err != nil {
		return fmt.Errorf("failed to update invoice ANAF status: %w", err)
	}

	return nil
}

// ProcessPendingANAFSubmissions processes invoices pending ANAF submission
func (s *InvoiceService) ProcessPendingANAFSubmissions(batchSize int) error {
	// Get pending invoices
	invoices, err := s.invoiceRepo.GetPendingANAFSubmission(batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending invoices: %w", err)
	}

	successCount := 0
	failCount := 0

	for _, invoice := range invoices {
		if err := s.SubmitToANAF(invoice.ID); err != nil {
			fmt.Printf("Failed to submit invoice %s to ANAF: %v\n", invoice.InvoiceNumber, err)
			failCount++
		} else {
			fmt.Printf("Successfully submitted invoice %s to ANAF\n", invoice.InvoiceNumber)
			successCount++
		}
	}

	fmt.Printf("ANAF submission batch complete: %d succeeded, %d failed\n", successCount, failCount)
	return nil
}

// ProcessANAFStatusUpdates checks status of invoices currently processing in ANAF
func (s *InvoiceService) ProcessANAFStatusUpdates() error {
	// Get invoices in processing status
	invoices, err := s.invoiceRepo.GetANAFProcessingInvoices()
	if err != nil {
		return fmt.Errorf("failed to get processing invoices: %w", err)
	}

	successCount := 0
	failCount := 0

	for _, invoice := range invoices {
		if err := s.CheckANAFStatus(invoice.ID); err != nil {
			fmt.Printf("Failed to check ANAF status for invoice %s: %v\n", invoice.InvoiceNumber, err)
			failCount++
		} else {
			fmt.Printf("Updated ANAF status for invoice %s: %s\n", invoice.InvoiceNumber, invoice.ANAFStatus)
			successCount++
		}
	}

	fmt.Printf("ANAF status check complete: %d succeeded, %d failed\n", successCount, failCount)
	return nil
}

// RetryFailedANAFSubmissions retries invoices that failed ANAF submission
func (s *InvoiceService) RetryFailedANAFSubmissions(maxRetries int) error {
	// Get failed invoices with retry count below max
	invoices, err := s.invoiceRepo.GetPendingANAFSubmission(10) // Get up to 10 failed invoices
	if err != nil {
		return fmt.Errorf("failed to get failed invoices: %w", err)
	}

	retryCount := 0
	for _, invoice := range invoices {
		if invoice.ANAFStatus == models.ANAFStatusFailed && invoice.ANAFRetryCount < maxRetries {
			if err := s.SubmitToANAF(invoice.ID); err != nil {
				fmt.Printf("Retry failed for invoice %s: %v\n", invoice.InvoiceNumber, err)
			} else {
				fmt.Printf("Retry successful for invoice %s\n", invoice.InvoiceNumber)
				retryCount++
			}
		}
	}

	fmt.Printf("Retried %d failed ANAF submissions\n", retryCount)
	return nil
}
