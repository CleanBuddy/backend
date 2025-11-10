package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// InvoiceStatus represents invoice status
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "DRAFT"
	InvoiceStatusIssued    InvoiceStatus = "ISSUED"
	InvoiceStatusPaid      InvoiceStatus = "PAID"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
)

// ANAFStatus represents ANAF submission status
type ANAFStatus string

const (
	ANAFStatusPending    ANAFStatus = "pending"    // Not yet submitted
	ANAFStatusProcessing ANAFStatus = "processing" // Submitted, awaiting ANAF response
	ANAFStatusAccepted   ANAFStatus = "accepted"   // Accepted by ANAF
	ANAFStatusRejected   ANAFStatus = "rejected"   // Rejected by ANAF (validation errors)
	ANAFStatusFailed     ANAFStatus = "failed"     // Technical failure (retry needed)
)

// ANAFError represents an error from ANAF
type ANAFError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Invoice represents an invoice for a booking
type Invoice struct {
	ID                 string
	BookingID          string
	InvoiceNumber      string
	IssueDate          time.Time
	DueDate            time.Time
	ClientName         string
	ClientEmail        sql.NullString
	CleanerName        string
	ServiceDescription string
	Subtotal           float64
	TaxAmount          float64
	TotalAmount        float64
	Currency           string
	Status             InvoiceStatus
	PdfURL             sql.NullString
	XmlURL             sql.NullString

	// ANAF e-Factura integration fields
	ANAFUploadIndex     sql.NullString
	ANAFStatus          ANAFStatus
	ANAFSubmittedAt     sql.NullTime
	ANAFProcessedAt     sql.NullTime
	ANAFDownloadID      sql.NullString
	ANAFConfirmationURL sql.NullString
	ANAFErrors          []ANAFError // Stored as JSONB in database
	ANAFRetryCount      int
	ANAFLastRetryAt     sql.NullTime

	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// InvoiceRepository handles invoice database operations
type InvoiceRepository struct {
	db *sql.DB
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// Create creates a new invoice
func (r *InvoiceRepository) Create(invoice *Invoice) error {
	// Set default ANAF status if not set
	if invoice.ANAFStatus == "" {
		invoice.ANAFStatus = ANAFStatusPending
	}

	return r.db.QueryRow(`
		INSERT INTO invoices (
			booking_id, invoice_number, issue_date, due_date,
			client_name, client_email, cleaner_name, service_description,
			subtotal, tax_amount, total_amount, currency, status,
			pdf_url, xml_url, anaf_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`, invoice.BookingID, invoice.InvoiceNumber, invoice.IssueDate, invoice.DueDate,
		invoice.ClientName, invoice.ClientEmail, invoice.CleanerName, invoice.ServiceDescription,
		invoice.Subtotal, invoice.TaxAmount, invoice.TotalAmount, invoice.Currency, invoice.Status,
		invoice.PdfURL, invoice.XmlURL, invoice.ANAFStatus).
		Scan(&invoice.ID, &invoice.CreatedAt, &invoice.UpdatedAt)
}

// GetByID finds an invoice by ID
func (r *InvoiceRepository) GetByID(id string) (*Invoice, error) {
	invoice := &Invoice{}
	var anafErrorsJSON sql.NullString

	err := r.db.QueryRow(`
		SELECT id, booking_id, invoice_number, issue_date, due_date,
		       client_name, client_email, cleaner_name, service_description,
		       subtotal, tax_amount, total_amount, currency, status,
		       pdf_url, xml_url,
		       anaf_upload_index, anaf_status, anaf_submitted_at, anaf_processed_at,
		       anaf_download_id, anaf_confirmation_url, anaf_errors,
		       anaf_retry_count, anaf_last_retry_at,
		       created_at, updated_at
		FROM invoices
		WHERE id = $1
	`, id).Scan(
		&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
		&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
		&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
		&invoice.PdfURL, &invoice.XmlURL,
		&invoice.ANAFUploadIndex, &invoice.ANAFStatus, &invoice.ANAFSubmittedAt, &invoice.ANAFProcessedAt,
		&invoice.ANAFDownloadID, &invoice.ANAFConfirmationURL, &anafErrorsJSON,
		&invoice.ANAFRetryCount, &invoice.ANAFLastRetryAt,
		&invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse ANAF errors JSON
	if anafErrorsJSON.Valid && anafErrorsJSON.String != "" {
		if err := json.Unmarshal([]byte(anafErrorsJSON.String), &invoice.ANAFErrors); err != nil {
			// Log error but don't fail the query
			fmt.Printf("Warning: failed to unmarshal ANAF errors: %v\n", err)
		}
	}

	return invoice, nil
}

// GetByBookingID finds an invoice by booking ID
func (r *InvoiceRepository) GetByBookingID(bookingID string) (*Invoice, error) {
	invoice := &Invoice{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, invoice_number, issue_date, due_date,
		       client_name, client_email, cleaner_name, service_description,
		       subtotal, tax_amount, total_amount, currency, status,
		       pdf_url, xml_url, created_at, updated_at
		FROM invoices
		WHERE booking_id = $1
	`, bookingID).Scan(
		&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
		&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
		&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
		&invoice.PdfURL, &invoice.XmlURL, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

// GetByInvoiceNumber finds an invoice by invoice number
func (r *InvoiceRepository) GetByInvoiceNumber(invoiceNumber string) (*Invoice, error) {
	invoice := &Invoice{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, invoice_number, issue_date, due_date,
		       client_name, client_email, cleaner_name, service_description,
		       subtotal, tax_amount, total_amount, currency, status,
		       pdf_url, xml_url, created_at, updated_at
		FROM invoices
		WHERE invoice_number = $1
	`, invoiceNumber).Scan(
		&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
		&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
		&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
		&invoice.PdfURL, &invoice.XmlURL, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

// GenerateInvoiceNumber generates next invoice number in format INV-YYYY-NNNN
func (r *InvoiceRepository) GenerateInvoiceNumber() (string, error) {
	var nextVal int64
	err := r.db.QueryRow("SELECT nextval('invoice_number_seq')").Scan(&nextVal)
	if err != nil {
		return "", fmt.Errorf("failed to generate invoice number: %w", err)
	}

	year := time.Now().Year()
	return fmt.Sprintf("INV-%d-%04d", year, nextVal), nil
}

// Update updates an invoice
func (r *InvoiceRepository) Update(invoice *Invoice) error {
	_, err := r.db.Exec(`
		UPDATE invoices
		SET invoice_number = $2, issue_date = $3, due_date = $4,
		    client_name = $5, client_email = $6, cleaner_name = $7, service_description = $8,
		    subtotal = $9, tax_amount = $10, total_amount = $11, currency = $12, status = $13,
		    pdf_url = $14, xml_url = $15, updated_at = NOW()
		WHERE id = $1
	`, invoice.ID, invoice.InvoiceNumber, invoice.IssueDate, invoice.DueDate,
		invoice.ClientName, invoice.ClientEmail, invoice.CleanerName, invoice.ServiceDescription,
		invoice.Subtotal, invoice.TaxAmount, invoice.TotalAmount, invoice.Currency, invoice.Status,
		invoice.PdfURL, invoice.XmlURL)
	return err
}

// GetByUserID returns all invoices for a user (either as client or cleaner)
func (r *InvoiceRepository) GetByUserID(userID string) ([]*Invoice, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.booking_id, i.invoice_number, i.issue_date, i.due_date,
		       i.client_name, i.client_email, i.cleaner_name, i.service_description,
		       i.subtotal, i.tax_amount, i.total_amount, i.currency, i.status,
		       i.pdf_url, i.xml_url, i.created_at, i.updated_at
		FROM invoices i
		JOIN bookings b ON i.booking_id = b.id
		WHERE b.client_id = $1 OR b.cleaner_id = $1
		ORDER BY i.issue_date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []*Invoice{}
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
			&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
			&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
			&invoice.PdfURL, &invoice.XmlURL, &invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}

// UpdateANAFStatus updates ANAF-related fields for an invoice
func (r *InvoiceRepository) UpdateANAFStatus(invoice *Invoice) error {
	// Serialize ANAF errors to JSON
	var anafErrorsJSON sql.NullString
	if len(invoice.ANAFErrors) > 0 {
		errorsBytes, err := json.Marshal(invoice.ANAFErrors)
		if err != nil {
			return fmt.Errorf("failed to marshal ANAF errors: %w", err)
		}
		anafErrorsJSON = sql.NullString{String: string(errorsBytes), Valid: true}
	}

	_, err := r.db.Exec(`
		UPDATE invoices
		SET anaf_upload_index = $2,
		    anaf_status = $3,
		    anaf_submitted_at = $4,
		    anaf_processed_at = $5,
		    anaf_download_id = $6,
		    anaf_confirmation_url = $7,
		    anaf_errors = $8,
		    anaf_retry_count = $9,
		    anaf_last_retry_at = $10,
		    updated_at = NOW()
		WHERE id = $1
	`, invoice.ID,
		invoice.ANAFUploadIndex,
		invoice.ANAFStatus,
		invoice.ANAFSubmittedAt,
		invoice.ANAFProcessedAt,
		invoice.ANAFDownloadID,
		invoice.ANAFConfirmationURL,
		anafErrorsJSON,
		invoice.ANAFRetryCount,
		invoice.ANAFLastRetryAt)
	return err
}

// GetPendingANAFSubmission returns invoices that need to be submitted to ANAF
func (r *InvoiceRepository) GetPendingANAFSubmission(limit int) ([]*Invoice, error) {
	rows, err := r.db.Query(`
		SELECT id, booking_id, invoice_number, issue_date, due_date,
		       client_name, client_email, cleaner_name, service_description,
		       subtotal, tax_amount, total_amount, currency, status,
		       pdf_url, xml_url,
		       anaf_upload_index, anaf_status, anaf_submitted_at, anaf_processed_at,
		       anaf_download_id, anaf_confirmation_url, anaf_errors,
		       anaf_retry_count, anaf_last_retry_at,
		       created_at, updated_at
		FROM invoices
		WHERE anaf_status IN ('pending', 'failed')
		  AND status = 'ISSUED'
		  AND (anaf_retry_count < 3 OR anaf_retry_count IS NULL)
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []*Invoice{}
	for rows.Next() {
		invoice := &Invoice{}
		var anafErrorsJSON sql.NullString

		err := rows.Scan(
			&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
			&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
			&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
			&invoice.PdfURL, &invoice.XmlURL,
			&invoice.ANAFUploadIndex, &invoice.ANAFStatus, &invoice.ANAFSubmittedAt, &invoice.ANAFProcessedAt,
			&invoice.ANAFDownloadID, &invoice.ANAFConfirmationURL, &anafErrorsJSON,
			&invoice.ANAFRetryCount, &invoice.ANAFLastRetryAt,
			&invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse ANAF errors JSON
		if anafErrorsJSON.Valid && anafErrorsJSON.String != "" {
			if err := json.Unmarshal([]byte(anafErrorsJSON.String), &invoice.ANAFErrors); err != nil {
				fmt.Printf("Warning: failed to unmarshal ANAF errors: %v\n", err)
			}
		}

		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}

// GetANAFProcessingInvoices returns invoices currently being processed by ANAF
func (r *InvoiceRepository) GetANAFProcessingInvoices() ([]*Invoice, error) {
	rows, err := r.db.Query(`
		SELECT id, booking_id, invoice_number, issue_date, due_date,
		       client_name, client_email, cleaner_name, service_description,
		       subtotal, tax_amount, total_amount, currency, status,
		       pdf_url, xml_url,
		       anaf_upload_index, anaf_status, anaf_submitted_at, anaf_processed_at,
		       anaf_download_id, anaf_confirmation_url, anaf_errors,
		       anaf_retry_count, anaf_last_retry_at,
		       created_at, updated_at
		FROM invoices
		WHERE anaf_status = 'processing'
		ORDER BY anaf_submitted_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []*Invoice{}
	for rows.Next() {
		invoice := &Invoice{}
		var anafErrorsJSON sql.NullString

		err := rows.Scan(
			&invoice.ID, &invoice.BookingID, &invoice.InvoiceNumber, &invoice.IssueDate, &invoice.DueDate,
			&invoice.ClientName, &invoice.ClientEmail, &invoice.CleanerName, &invoice.ServiceDescription,
			&invoice.Subtotal, &invoice.TaxAmount, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
			&invoice.PdfURL, &invoice.XmlURL,
			&invoice.ANAFUploadIndex, &invoice.ANAFStatus, &invoice.ANAFSubmittedAt, &invoice.ANAFProcessedAt,
			&invoice.ANAFDownloadID, &invoice.ANAFConfirmationURL, &anafErrorsJSON,
			&invoice.ANAFRetryCount, &invoice.ANAFLastRetryAt,
			&invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse ANAF errors JSON
		if anafErrorsJSON.Valid && anafErrorsJSON.String != "" {
			if err := json.Unmarshal([]byte(anafErrorsJSON.String), &invoice.ANAFErrors); err != nil {
				fmt.Printf("Warning: failed to unmarshal ANAF errors: %v\n", err)
			}
		}

		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}
