package services

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/models"
)

// XMLGenerator handles UBL 2.1 XML generation for ANAF e-Factura
type XMLGenerator struct {
	outputDir string
	config    *config.CompanyConfig
}

// NewXMLGenerator creates a new XML generator
func NewXMLGenerator(outputDir string, companyConfig *config.CompanyConfig) *XMLGenerator {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create XML output directory: %v\n", err)
	}

	return &XMLGenerator{
		outputDir: outputDir,
		config:    companyConfig,
	}
}

// UBL Invoice structure (simplified UBL 2.1)
type UBLInvoice struct {
	XMLName xml.Name `xml:"Invoice"`
	XMLNS   string   `xml:"xmlns,attr"`
	CAC     string   `xml:"xmlns:cac,attr"`
	CBC     string   `xml:"xmlns:cbc,attr"`

	CustomizationID string         `xml:"cbc:CustomizationID"`
	ID              string         `xml:"cbc:ID"`
	IssueDate       string         `xml:"cbc:IssueDate"`
	DueDate         string         `xml:"cbc:DueDate"`
	InvoiceTypeCode string         `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode string    `xml:"cbc:DocumentCurrencyCode"`

	AccountingSupplierParty UBLParty `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty UBLParty `xml:"cac:AccountingCustomerParty"`

	TaxTotal           UBLTaxTotal       `xml:"cac:TaxTotal"`
	LegalMonetaryTotal UBLMonetaryTotal  `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines       []UBLInvoiceLine  `xml:"cac:InvoiceLine"`
	PaymentMeans       UBLPaymentMeans   `xml:"cac:PaymentMeans"`
}

type UBLParty struct {
	Party struct {
		PartyName struct {
			Name string `xml:"cbc:Name"`
		} `xml:"cac:PartyName"`
		PostalAddress struct {
			Country struct {
				IdentificationCode string `xml:"cbc:IdentificationCode"`
			} `xml:"cac:Country"`
		} `xml:"cac:PostalAddress"`
		PartyTaxScheme struct {
			CompanyID string `xml:"cbc:CompanyID"`
			TaxScheme struct {
				ID string `xml:"cbc:ID"`
			} `xml:"cac:TaxScheme"`
		} `xml:"cac:PartyTaxScheme"`
		Contact struct {
			ElectronicMail string `xml:"cbc:ElectronicMail,omitempty"`
		} `xml:"cac:Contact"`
	} `xml:"cac:Party"`
}

type UBLTaxTotal struct {
	TaxAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:TaxAmount"`
	TaxSubtotal struct {
		TaxableAmount struct {
			Value    float64 `xml:",chardata"`
			Currency string  `xml:"currencyID,attr"`
		} `xml:"cbc:TaxableAmount"`
		TaxAmount struct {
			Value    float64 `xml:",chardata"`
			Currency string  `xml:"currencyID,attr"`
		} `xml:"cbc:TaxAmount"`
		TaxCategory struct {
			ID      string `xml:"cbc:ID"`
			Percent string `xml:"cbc:Percent"`
			TaxScheme struct {
				ID string `xml:"cbc:ID"`
			} `xml:"cac:TaxScheme"`
		} `xml:"cac:TaxCategory"`
	} `xml:"cac:TaxSubtotal"`
}

type UBLMonetaryTotal struct {
	LineExtensionAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:LineExtensionAmount"`
	TaxExclusiveAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:TaxExclusiveAmount"`
	TaxInclusiveAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:PayableAmount"`
}

type UBLInvoiceLine struct {
	ID string `xml:"cbc:ID"`
	InvoicedQuantity struct {
		Value float64 `xml:",chardata"`
		Unit  string  `xml:"unitCode,attr"`
	} `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount struct {
		Value    float64 `xml:",chardata"`
		Currency string  `xml:"currencyID,attr"`
	} `xml:"cbc:LineExtensionAmount"`
	Item struct {
		Description string `xml:"cbc:Description"`
		Name        string `xml:"cbc:Name"`
	} `xml:"cac:Item"`
	Price struct {
		PriceAmount struct {
			Value    float64 `xml:",chardata"`
			Currency string  `xml:"currencyID,attr"`
		} `xml:"cbc:PriceAmount"`
	} `xml:"cac:Price"`
}

type UBLPaymentMeans struct {
	PaymentMeansCode      string `xml:"cbc:PaymentMeansCode"`
	PayeeFinancialAccount struct {
		ID   string `xml:"cbc:ID"`
		Name string `xml:"cbc:Name"`
		FinancialInstitutionBranch struct {
			ID string `xml:"cbc:ID"`
		} `xml:"cac:FinancialInstitutionBranch"`
	} `xml:"cac:PayeeFinancialAccount"`
}

// GenerateInvoiceXML generates UBL 2.1 XML for ANAF e-Factura
func (g *XMLGenerator) GenerateInvoiceXML(invoice *models.Invoice, booking *models.Booking) (string, error) {
	// Calculate VAT breakdown
	// If invoice already has VAT calculated, use it; otherwise calculate from total
	var subtotal, vatAmount float64
	if invoice.TaxAmount > 0 {
		// VAT already calculated
		subtotal = invoice.Subtotal
		vatAmount = invoice.TaxAmount
	} else {
		// Calculate VAT from total (assuming total includes 19% VAT)
		vatRate := g.config.VATRate
		if vatRate == 0 {
			vatRate = 0.19 // Default 19% VAT for Romania
		}
		subtotal = invoice.TotalAmount / (1 + vatRate)
		vatAmount = invoice.TotalAmount - subtotal
	}

	// Create UBL structure
	ubl := UBLInvoice{
		XMLNS:           "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		CAC:             "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		CBC:             "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		CustomizationID: "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0",
		ID:              invoice.InvoiceNumber,
		IssueDate:       invoice.IssueDate.Format("2006-01-02"),
		DueDate:         invoice.DueDate.Format("2006-01-02"),
		InvoiceTypeCode: "380", // Commercial invoice
		DocumentCurrencyCode: invoice.Currency,
	}

	// Supplier (CleanBuddy) - Use config values
	ubl.AccountingSupplierParty.Party.PartyName.Name = g.config.LegalName
	ubl.AccountingSupplierParty.Party.PostalAddress.Country.IdentificationCode = g.config.Address.Country
	ubl.AccountingSupplierParty.Party.PartyTaxScheme.CompanyID = g.config.CUI
	ubl.AccountingSupplierParty.Party.PartyTaxScheme.TaxScheme.ID = "VAT"
	ubl.AccountingSupplierParty.Party.Contact.ElectronicMail = g.config.Contact.Email

	// Customer
	ubl.AccountingCustomerParty.Party.PartyName.Name = invoice.ClientName
	ubl.AccountingCustomerParty.Party.PostalAddress.Country.IdentificationCode = "RO"
	if invoice.ClientEmail.Valid {
		ubl.AccountingCustomerParty.Party.Contact.ElectronicMail = invoice.ClientEmail.String
	}

	// Tax breakdown
	vatPercent := g.config.VATRate * 100
	if vatPercent == 0 {
		vatPercent = 19 // Default 19%
	}
	ubl.TaxTotal.TaxAmount.Value = vatAmount
	ubl.TaxTotal.TaxAmount.Currency = invoice.Currency
	ubl.TaxTotal.TaxSubtotal.TaxableAmount.Value = subtotal
	ubl.TaxTotal.TaxSubtotal.TaxableAmount.Currency = invoice.Currency
	ubl.TaxTotal.TaxSubtotal.TaxAmount.Value = vatAmount
	ubl.TaxTotal.TaxSubtotal.TaxAmount.Currency = invoice.Currency
	ubl.TaxTotal.TaxSubtotal.TaxCategory.ID = "S" // Standard rate
	ubl.TaxTotal.TaxSubtotal.TaxCategory.Percent = fmt.Sprintf("%.0f", vatPercent)
	ubl.TaxTotal.TaxSubtotal.TaxCategory.TaxScheme.ID = "VAT"

	// Monetary totals
	ubl.LegalMonetaryTotal.LineExtensionAmount.Value = subtotal
	ubl.LegalMonetaryTotal.LineExtensionAmount.Currency = invoice.Currency
	ubl.LegalMonetaryTotal.TaxExclusiveAmount.Value = subtotal
	ubl.LegalMonetaryTotal.TaxExclusiveAmount.Currency = invoice.Currency
	ubl.LegalMonetaryTotal.TaxInclusiveAmount.Value = invoice.TotalAmount
	ubl.LegalMonetaryTotal.TaxInclusiveAmount.Currency = invoice.Currency
	ubl.LegalMonetaryTotal.PayableAmount.Value = invoice.TotalAmount
	ubl.LegalMonetaryTotal.PayableAmount.Currency = invoice.Currency

	// Payment information
	ubl.PaymentMeans.PaymentMeansCode = "30" // Credit transfer
	ubl.PaymentMeans.PayeeFinancialAccount.ID = g.config.Bank.IBAN
	ubl.PaymentMeans.PayeeFinancialAccount.Name = g.config.LegalName
	ubl.PaymentMeans.PayeeFinancialAccount.FinancialInstitutionBranch.ID = g.config.Bank.SWIFT

	// Generate detailed invoice lines from booking
	ubl.InvoiceLines = g.generateInvoiceLines(booking, subtotal, invoice.Currency)

	// Marshal to XML
	output, err := xml.MarshalIndent(ubl, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header
	xmlContent := xml.Header + string(output)

	// Generate filename
	filename := fmt.Sprintf("invoice_%s_%d.xml",
		invoice.InvoiceNumber,
		time.Now().Unix())
	filepath := filepath.Join(g.outputDir, filename)

	// Write to file
	if err := os.WriteFile(filepath, []byte(xmlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write XML file: %w", err)
	}

	return filepath, nil
}

// generateInvoiceLines creates detailed invoice lines from booking information
func (g *XMLGenerator) generateInvoiceLines(booking *models.Booking, totalSubtotal float64, currency string) []UBLInvoiceLine {
	if booking == nil {
		// Fallback to single line if no booking data
		line := UBLInvoiceLine{ID: "1"}
		line.InvoicedQuantity.Value = 1
		line.InvoicedQuantity.Unit = "HUR"
		line.LineExtensionAmount.Value = totalSubtotal
		line.LineExtensionAmount.Currency = currency
		line.Item.Name = "Servicii de curățenie profesionale"
		line.Item.Description = "Servicii de curățenie"
		line.Price.PriceAmount.Value = totalSubtotal
		line.Price.PriceAmount.Currency = currency
		return []UBLInvoiceLine{line}
	}

	lines := []UBLInvoiceLine{}
	lineID := 1

	// Calculate base service price (exclude add-ons from base price)
	basePrice := totalSubtotal
	windowPrice := 0.0
	carpetPrice := 0.0

	// Estimate add-on prices if included
	if booking.IncludesWindows {
		windowPrice = float64(booking.NumberOfWindows) * 10 // 10 RON per window estimate
		basePrice -= windowPrice
	}
	if booking.IncludesCarpetCleaning {
		carpetPrice = float64(booking.CarpetAreaSqm) * 5 // 5 RON per sqm estimate
		basePrice -= carpetPrice
	}

	// Ensure base price is not negative
	if basePrice < 0 {
		basePrice = totalSubtotal * 0.8 // Default 80% for base service
		if booking.IncludesWindows {
			windowPrice = totalSubtotal * 0.1
		}
		if booking.IncludesCarpetCleaning {
			carpetPrice = totalSubtotal * 0.1
		}
	}

	// Line 1: Base cleaning service
	baseLine := UBLInvoiceLine{ID: fmt.Sprintf("%d", lineID)}
	baseLine.InvoicedQuantity.Value = float64(booking.EstimatedHours)
	baseLine.InvoicedQuantity.Unit = "HUR" // Hours
	baseLine.LineExtensionAmount.Value = basePrice
	baseLine.LineExtensionAmount.Currency = currency

	// Build service name
	serviceName := g.translateServiceType(booking.ServiceType)
	if booking.IncludesDeepCleaning {
		serviceName += " cu curățenie profundă"
	}
	baseLine.Item.Name = serviceName

	// Build description
	description := fmt.Sprintf("Serviciu de curățenie - %s", serviceName)
	if booking.AreaSqm.Valid {
		description += fmt.Sprintf(", Suprafață: %d mp", booking.AreaSqm.Int32)
	}
	description += fmt.Sprintf(", Durata: %d ore", booking.EstimatedHours)
	description += fmt.Sprintf(", Data: %s", booking.ScheduledDate.Format("02.01.2006"))
	baseLine.Item.Description = description

	baseLine.Price.PriceAmount.Value = basePrice / float64(booking.EstimatedHours)
	baseLine.Price.PriceAmount.Currency = currency
	lines = append(lines, baseLine)
	lineID++

	// Line 2: Window cleaning (if applicable)
	if booking.IncludesWindows && booking.NumberOfWindows > 0 {
		windowLine := UBLInvoiceLine{ID: fmt.Sprintf("%d", lineID)}
		windowLine.InvoicedQuantity.Value = float64(booking.NumberOfWindows)
		windowLine.InvoicedQuantity.Unit = "C62" // Unit (pieces)
		windowLine.LineExtensionAmount.Value = windowPrice
		windowLine.LineExtensionAmount.Currency = currency
		windowLine.Item.Name = "Curățare geamuri"
		windowLine.Item.Description = fmt.Sprintf("Curățare geamuri - %d bucăți", booking.NumberOfWindows)
		windowLine.Price.PriceAmount.Value = windowPrice / float64(booking.NumberOfWindows)
		windowLine.Price.PriceAmount.Currency = currency
		lines = append(lines, windowLine)
		lineID++
	}

	// Line 3: Carpet cleaning (if applicable)
	if booking.IncludesCarpetCleaning && booking.CarpetAreaSqm > 0 {
		carpetLine := UBLInvoiceLine{ID: fmt.Sprintf("%d", lineID)}
		carpetLine.InvoicedQuantity.Value = float64(booking.CarpetAreaSqm)
		carpetLine.InvoicedQuantity.Unit = "MTK" // Square meters
		carpetLine.LineExtensionAmount.Value = carpetPrice
		carpetLine.LineExtensionAmount.Currency = currency
		carpetLine.Item.Name = "Curățare covoare/mochetă"
		carpetLine.Item.Description = fmt.Sprintf("Curățare covoare/mochetă - %d mp", booking.CarpetAreaSqm)
		carpetLine.Price.PriceAmount.Value = carpetPrice / float64(booking.CarpetAreaSqm)
		carpetLine.Price.PriceAmount.Currency = currency
		lines = append(lines, carpetLine)
	}

	return lines
}

// translateServiceType translates service type enum to Romanian
func (g *XMLGenerator) translateServiceType(serviceType models.ServiceType) string {
	switch serviceType {
	case models.ServiceTypeStandard:
		return "Curățenie Standard"
	case models.ServiceTypeDeepCleaning:
		return "Curățenie Profundă"
	case models.ServiceTypeOffice:
		return "Curățenie Birou"
	case models.ServiceTypePostRenovation:
		return "Curățenie Post-Renovare"
	case models.ServiceTypeMoveInOut:
		return "Curățenie Mutare"
	default:
		return "Servicii de curățenie"
	}
}

// ReadXML reads an XML file from disk
func (g *XMLGenerator) ReadXML(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
