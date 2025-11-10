package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/jung-kurt/gofpdf"
)

// PDFGenerator handles invoice PDF generation
type PDFGenerator struct {
	outputDir string
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator(outputDir string) *PDFGenerator {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create PDF output directory: %v\n", err)
	}

	return &PDFGenerator{
		outputDir: outputDir,
	}
}

// GenerateInvoicePDF generates a PDF invoice and returns the file path
func (g *PDFGenerator) GenerateInvoicePDF(invoice *models.Invoice) (string, error) {
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set colors
	primaryColor := struct{ R, G, B int }{16, 185, 129} // Green-500

	// Header - Company Logo Area
	pdf.SetFillColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.Rect(0, 0, 210, 40, "F")

	// Company Name
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetXY(15, 12)
	pdf.Cell(0, 10, "CleanBuddy")

	// Company tagline
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(15, 24)
	pdf.Cell(0, 5, "Servicii Profesionale de Curatenie")

	// Reset text color
	pdf.SetTextColor(0, 0, 0)

	// Invoice Title
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(15, 50)
	pdf.Cell(0, 8, "FACTURA")

	// Invoice Number and Date
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(150, 50)
	pdf.Cell(0, 5, fmt.Sprintf("Nr: %s", invoice.InvoiceNumber))
	pdf.SetXY(150, 56)
	pdf.Cell(0, 5, fmt.Sprintf("Data: %s", invoice.IssueDate.Format("02.01.2006")))
	pdf.SetXY(150, 62)
	pdf.Cell(0, 5, fmt.Sprintf("Scadenta: %s", invoice.DueDate.Format("02.01.2006")))

	// Client Information
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(15, 75)
	pdf.Cell(0, 6, "Catre:")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(15, 82)
	pdf.Cell(0, 5, invoice.ClientName)

	if invoice.ClientEmail.Valid {
		pdf.SetXY(15, 88)
		pdf.Cell(0, 5, invoice.ClientEmail.String)
	}

	// Provider Information (Right side)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(120, 75)
	pdf.Cell(0, 6, "De la:")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(120, 82)
	pdf.Cell(0, 5, "CleanBuddy SRL")
	pdf.SetXY(120, 88)
	pdf.Cell(0, 5, "CUI: RO12345678")
	pdf.SetXY(120, 94)
	pdf.Cell(0, 5, "Bucuresti, Romania")

	// Table Header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)

	startY := 115.0
	pdf.SetXY(15, startY)
	pdf.CellFormat(90, 8, "Descriere Serviciu", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "Cant.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Pret Unitar", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 1, "R", true, 0, "")

	// Table Row - Service
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(15, startY+8)
	pdf.CellFormat(90, 8, invoice.ServiceDescription, "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 8, "1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f %s", invoice.Subtotal, invoice.Currency), "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f %s", invoice.Subtotal, invoice.Currency), "1", 1, "R", false, 0, "")

	// Subtotal
	summaryY := startY + 20
	pdf.SetXY(125, summaryY)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(35, 6, "Subtotal:")
	pdf.SetXY(160, summaryY)
	pdf.CellFormat(30, 6, fmt.Sprintf("%.2f %s", invoice.Subtotal, invoice.Currency), "", 1, "R", false, 0, "")

	// Tax (TVA)
	if invoice.TaxAmount > 0 {
		pdf.SetXY(125, summaryY+6)
		pdf.Cell(35, 6, "TVA (19%):")
		pdf.SetXY(160, summaryY+6)
		pdf.CellFormat(30, 6, fmt.Sprintf("%.2f %s", invoice.TaxAmount, invoice.Currency), "", 1, "R", false, 0, "")
		summaryY += 6
	}

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(125, summaryY+6)
	pdf.Cell(35, 7, "TOTAL:")
	pdf.SetXY(160, summaryY+6)
	pdf.CellFormat(30, 7, fmt.Sprintf("%.2f %s", invoice.TotalAmount, invoice.Currency), "", 1, "R", false, 0, "")

	// Cleaner Information
	pdf.SetFont("Arial", "I", 9)
	pdf.SetXY(15, summaryY+20)
	pdf.Cell(0, 5, fmt.Sprintf("Serviciu executat de: %s", invoice.CleanerName))

	// Payment Information
	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(15, 250)
	pdf.Cell(0, 5, "Modalitate plata: Card / Transfer bancar")
	pdf.SetXY(15, 256)
	pdf.Cell(0, 5, "Cont bancar: RO49AAAA1B31007593840000 (BCR)")

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.SetXY(15, 280)
	pdf.Cell(0, 4, "Multumim pentru incredere!")
	pdf.SetXY(15, 285)
	pdf.Cell(0, 4, "CleanBuddy - Servicii profesionale de curatenie")
	pdf.SetXY(15, 290)
	pdf.Cell(0, 4, "contact@cleanbuddy.ro | +40 123 456 789")

	// Generate filename
	filename := fmt.Sprintf("invoice_%s_%d.pdf",
		invoice.InvoiceNumber,
		time.Now().Unix())
	filepath := filepath.Join(g.outputDir, filename)

	// Save PDF
	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return filepath, nil
}
