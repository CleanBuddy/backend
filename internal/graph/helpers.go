package graph

import (
	"log"
	"time"

	"github.com/cleanbuddy/backend/internal/graph/model"
	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// convertUserToGraphQL converts database user model to GraphQL model
func convertUserToGraphQL(user *models.User) *model.User {
	var email, phone, firstName, lastName *string

	if user.Email.Valid {
		email = &user.Email.String
	}
	if user.Phone.Valid {
		phone = &user.Phone.String
	}
	if user.FirstName.Valid {
		firstName = &user.FirstName.String
	}
	if user.LastName.Valid {
		lastName = &user.LastName.String
	}

	return &model.User{
		ID:        user.ID,
		Email:     email,
		Phone:     phone,
		FirstName: firstName,
		LastName:  lastName,
		Role:      model.UserRole(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// convertClientToGraphQL converts database client model to GraphQL model
func convertClientToGraphQL(client *models.Client) *model.Client {
	var phoneNumber *string
	var averageRating *float64

	if client.PhoneNumber.Valid {
		phoneNumber = &client.PhoneNumber.String
	}
	if client.AverageRating.Valid {
		averageRating = &client.AverageRating.Float64
	}

	return &model.Client{
		ID:                client.ID,
		UserID:            client.UserID,
		PhoneNumber:       phoneNumber,
		PreferredLanguage: client.PreferredLanguage,
		TotalBookings:     client.TotalBookings,
		TotalSpent:        client.TotalSpent,
		AverageRating:     averageRating,
		CreatedAt:         client.CreatedAt,
		UpdatedAt:         client.UpdatedAt,
	}
}

// convertAddressToGraphQL converts database address model to GraphQL model
func convertAddressToGraphQL(address *models.Address) *model.Address {
	var apartment, postalCode, additionalInfo *string

	if address.Apartment.Valid {
		apartment = &address.Apartment.String
	}
	if address.PostalCode.Valid {
		postalCode = &address.PostalCode.String
	}
	if address.AdditionalInfo.Valid {
		additionalInfo = &address.AdditionalInfo.String
	}

	return &model.Address{
		ID:             address.ID,
		UserID:         address.UserID,
		Label:          address.Label,
		StreetAddress:  address.StreetAddress,
		Apartment:      apartment,
		City:           address.City,
		County:         address.County,
		PostalCode:     postalCode,
		Country:        address.Country,
		AdditionalInfo: additionalInfo,
		IsDefault:      address.IsDefault,
		CreatedAt:      address.CreatedAt,
		UpdatedAt:      address.UpdatedAt,
	}
}

// convertCleanerToGraphQL converts database cleaner model to GraphQL model
func convertCleanerToGraphQL(cleaner *models.Cleaner) *model.Cleaner {
	var dateOfBirth, streetAddress, city, county, postalCode, bio *string
	var idDocumentURL, backgroundCheckURL, profilePhotoURL *string
	var iban *string
	var averageRating *float64

	if cleaner.DateOfBirth.Valid {
		dob := cleaner.DateOfBirth.Time.Format("2006-01-02")
		dateOfBirth = &dob
	}
	if cleaner.StreetAddress.Valid {
		streetAddress = &cleaner.StreetAddress.String
	}
	if cleaner.City.Valid {
		city = &cleaner.City.String
	}
	if cleaner.County.Valid {
		county = &cleaner.County.String
	}
	if cleaner.PostalCode.Valid {
		postalCode = &cleaner.PostalCode.String
	}
	if cleaner.Bio.Valid {
		bio = &cleaner.Bio.String
	}
	if cleaner.IBAN.Valid {
		// Decrypt IBAN before returning to GraphQL
		decryptedIBAN, err := utils.DecryptIBAN(cleaner.IBAN.String)
		if err != nil {
			// Log error but don't fail - return nil IBAN instead
			log.Printf("Warning: Failed to decrypt IBAN for cleaner %s: %v", cleaner.ID, err)
		} else if decryptedIBAN != "" {
			iban = &decryptedIBAN
		}
	}
	if cleaner.IDDocumentURL.Valid {
		idDocumentURL = &cleaner.IDDocumentURL.String
	}
	if cleaner.BackgroundCheckURL.Valid {
		backgroundCheckURL = &cleaner.BackgroundCheckURL.String
	}
	if cleaner.ProfilePhotoURL.Valid {
		profilePhotoURL = &cleaner.ProfilePhotoURL.String
	}
	if cleaner.AverageRating.Valid {
		averageRating = &cleaner.AverageRating.Float64
	}

	// Parse JSONB arrays using helper methods
	specializations, err := cleaner.ParseSpecializations()
	if err != nil {
		log.Printf("Warning: Failed to parse specializations for cleaner %s: %v", cleaner.ID, err)
		specializations = []string{} // Default to empty array on error
	}

	languages, err := cleaner.ParseLanguages()
	if err != nil {
		log.Printf("Warning: Failed to parse languages for cleaner %s: %v", cleaner.ID, err)
		languages = []string{"ro"} // Default to Romanian on error
	}
	if len(languages) == 0 {
		languages = []string{"ro"} // Default to Romanian if empty
	}

	return &model.Cleaner{
		ID:                      cleaner.ID,
		UserID:                  cleaner.UserID,
		PhoneNumber:             cleaner.PhoneNumber,
		DateOfBirth:             dateOfBirth,
		StreetAddress:           streetAddress,
		City:                    city,
		County:                  county,
		PostalCode:              postalCode,
		YearsOfExperience:       cleaner.YearsOfExperience,
		Bio:                     bio,
		Specializations:         specializations,
		Languages:               languages,
		Iban:                    iban,
		IDDocumentURL:           idDocumentURL,
		IDDocumentVerified:      cleaner.IDDocumentVerified,
		BackgroundCheckURL:      backgroundCheckURL,
		BackgroundCheckVerified: cleaner.BackgroundCheckVerified,
		ProfilePhotoURL:         profilePhotoURL,
		AverageRating:           averageRating,
		TotalJobs:               cleaner.TotalJobs,
		TotalEarnings:           cleaner.TotalEarnings,
		ApprovalStatus:          model.ApprovalStatus(cleaner.ApprovalStatus),
		IsActive:                cleaner.IsActive,
		IsAvailable:             cleaner.IsAvailable,
		CreatedAt:               cleaner.CreatedAt,
		UpdatedAt:               cleaner.UpdatedAt,
	}
}

// convertBookingToGraphQL converts database booking model to GraphQL model
func convertBookingToGraphQL(booking *models.Booking) *model.Booking {
	var cleanerID, specialInstructions, accessInstructions, reservationCode, timePreferences *string
	var scheduledDate, scheduledTime *time.Time
	var confirmedAt, startedAt, completedAt, cancelledAt *time.Time
	var cancelledBy, cancellationReason *string
	var clientRating, cleanerRating *int
	var clientReview, cleanerReview *string
	var areaSqm *int

	if booking.CleanerID.Valid {
		cleanerID = &booking.CleanerID.String
	}
	if booking.ReservationCode.Valid {
		reservationCode = &booking.ReservationCode.String
	}
	if booking.SpecialInstructions.Valid {
		specialInstructions = &booking.SpecialInstructions.String
	}
	if booking.AccessInstructions.Valid {
		accessInstructions = &booking.AccessInstructions.String
	}
	if booking.TimePreferences.Valid {
		timePreferences = &booking.TimePreferences.String
	}
	if !booking.ScheduledDate.IsZero() {
		scheduledDate = &booking.ScheduledDate
	}
	if !booking.ScheduledTime.IsZero() {
		scheduledTime = &booking.ScheduledTime
	}
	if booking.ConfirmedAt.Valid {
		confirmedAt = &booking.ConfirmedAt.Time
	}
	if booking.StartedAt.Valid {
		startedAt = &booking.StartedAt.Time
	}
	if booking.CompletedAt.Valid {
		completedAt = &booking.CompletedAt.Time
	}
	if booking.CancelledAt.Valid {
		cancelledAt = &booking.CancelledAt.Time
	}
	if booking.CancelledBy.Valid {
		cancelledBy = &booking.CancelledBy.String
	}
	if booking.CancellationReason.Valid {
		cancellationReason = &booking.CancellationReason.String
	}
	if booking.ClientRating.Valid {
		rating := int(booking.ClientRating.Int32)
		clientRating = &rating
	}
	if booking.ClientReview.Valid {
		clientReview = &booking.ClientReview.String
	}
	if booking.CleanerRating.Valid {
		rating := int(booking.CleanerRating.Int32)
		cleanerRating = &rating
	}
	if booking.CleanerReview.Valid {
		cleanerReview = &booking.CleanerReview.String
	}
	if booking.AreaSqm.Valid {
		sqm := int(booking.AreaSqm.Int32)
		areaSqm = &sqm
	}

	return &model.Booking{
		ID:                     booking.ID,
		ReservationCode:        reservationCode,
		ClientID:               booking.ClientID,
		CleanerID:              cleanerID,
		AddressID:              booking.AddressID,
		ServiceType:            model.ServiceType(booking.ServiceType),
		AreaSqm:                areaSqm,
		EstimatedHours:         booking.EstimatedHours,
		ScheduledDate:          scheduledDate,
		ScheduledTime:          scheduledTime,
		TimePreferences:        timePreferences,
		IncludesDeepCleaning:   booking.IncludesDeepCleaning,
		IncludesWindows:        booking.IncludesWindows,
		IncludesCarpetCleaning: booking.IncludesCarpetCleaning,
		IncludesFridge:         booking.IncludesFridgeCleaning,
		IncludesOven:           booking.IncludesOvenCleaning,
		IncludesBalcony:        booking.IncludesBalconyCleaning,
		NumberOfWindows:        booking.NumberOfWindows,
		CarpetAreaSqm:          booking.CarpetAreaSqm,
		BasePrice:              booking.BasePrice,
		AddonsPrice:            booking.AddonsPrice,
		TotalPrice:             booking.TotalPrice,
		PlatformFee:            booking.PlatformFee,
		CleanerPayout:          booking.CleanerPayout,
		DiscountApplied:        booking.DiscountApplied,
		Status:                 model.BookingStatus(booking.Status),
		SpecialInstructions:    specialInstructions,
		AccessInstructions:     accessInstructions,
		ConfirmedAt:            confirmedAt,
		StartedAt:              startedAt,
		CompletedAt:            completedAt,
		CancelledAt:            cancelledAt,
		CancelledBy:            cancelledBy,
		CancellationReason:     cancellationReason,
		ClientRating:           clientRating,
		ClientReview:           clientReview,
		CleanerRating:          cleanerRating,
		CleanerReview:          cleanerReview,
		CreatedAt:              booking.CreatedAt,
		UpdatedAt:              booking.UpdatedAt,
	}
}

// convertPaymentToGraphQL converts database payment model to GraphQL model
func convertPaymentToGraphQL(payment *models.Payment) *model.Payment {
	var providerTransactionID, providerOrderID, cardLastFour, cardBrand *string
	var errorCode, errorMessage *string
	var authorizedAt, capturedAt, failedAt, refundedAt *time.Time

	if payment.ProviderTransactionID.Valid {
		providerTransactionID = &payment.ProviderTransactionID.String
	}
	if payment.ProviderOrderID.Valid {
		providerOrderID = &payment.ProviderOrderID.String
	}
	if payment.CardLastFour.Valid {
		cardLastFour = &payment.CardLastFour.String
	}
	if payment.CardBrand.Valid {
		cardBrand = &payment.CardBrand.String
	}
	if payment.ErrorCode.Valid {
		errorCode = &payment.ErrorCode.String
	}
	if payment.ErrorMessage.Valid {
		errorMessage = &payment.ErrorMessage.String
	}
	if payment.AuthorizedAt.Valid {
		authorizedAt = &payment.AuthorizedAt.Time
	}
	if payment.CapturedAt.Valid {
		capturedAt = &payment.CapturedAt.Time
	}
	if payment.FailedAt.Valid {
		failedAt = &payment.FailedAt.Time
	}
	if payment.RefundedAt.Valid {
		refundedAt = &payment.RefundedAt.Time
	}

	return &model.Payment{
		ID:                    payment.ID,
		BookingID:             payment.BookingID,
		UserID:                payment.UserID,
		Provider:              model.PaymentProvider(payment.Provider),
		ProviderTransactionID: providerTransactionID,
		ProviderOrderID:       providerOrderID,
		PaymentType:           model.PaymentType(payment.PaymentType),
		Status:                model.PaymentStatus(payment.Status),
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		CardLastFour:          cardLastFour,
		CardBrand:             cardBrand,
		ErrorCode:             errorCode,
		ErrorMessage:          errorMessage,
		AuthorizedAt:          authorizedAt,
		CapturedAt:            capturedAt,
		FailedAt:              failedAt,
		RefundedAt:            refundedAt,
		CreatedAt:             payment.CreatedAt,
		UpdatedAt:             payment.UpdatedAt,
	}
}

// convertAvailabilityToGraphQL converts database availability model to GraphQL model
func convertAvailabilityToGraphQL(availability *models.Availability) *model.Availability {
	var dayOfWeek *int
	var specificDate *time.Time
	var notes *string

	if availability.DayOfWeek.Valid {
		dow := int(availability.DayOfWeek.Int32)
		dayOfWeek = &dow
	}
	if availability.SpecificDate.Valid {
		specificDate = &availability.SpecificDate.Time
	}
	if availability.Notes.Valid {
		notes = &availability.Notes.String
	}

	return &model.Availability{
		ID:           availability.ID,
		CleanerID:    availability.CleanerID,
		Type:         model.AvailabilityType(availability.Type),
		DayOfWeek:    dayOfWeek,
		SpecificDate: specificDate,
		StartTime:    availability.StartTime,
		EndTime:      availability.EndTime,
		IsActive:     availability.IsActive,
		Notes:        notes,
		CreatedAt:    availability.CreatedAt,
		UpdatedAt:    availability.UpdatedAt,
	}
}

// convertCompanyToGraphQL converts database company model to GraphQL model
func convertCompanyToGraphQL(company *models.Company) *model.Company {
	var registrationNumber, iban, bankName, legalAddress *string
	var contactEmail, contactPhone, rejectedReason *string

	if company.RegistrationNumber.Valid {
		registrationNumber = &company.RegistrationNumber.String
	}
	if company.IBAN.Valid {
		// Decrypt IBAN before returning to GraphQL
		decryptedIBAN, err := utils.DecryptIBAN(company.IBAN.String)
		if err != nil {
			// Log error but don't fail - return nil IBAN instead
			log.Printf("Warning: Failed to decrypt IBAN for company %s: %v", company.ID, err)
		} else if decryptedIBAN != "" {
			iban = &decryptedIBAN
		}
	}
	if company.BankName.Valid {
		bankName = &company.BankName.String
	}
	if company.LegalAddress.Valid {
		legalAddress = &company.LegalAddress.String
	}
	if company.ContactEmail.Valid {
		contactEmail = &company.ContactEmail.String
	}
	if company.ContactPhone.Valid {
		contactPhone = &company.ContactPhone.String
	}
	if company.RejectedReason.Valid {
		rejectedReason = &company.RejectedReason.String
	}

	return &model.Company{
		ID:                 company.ID,
		Name:               company.Name,
		Cui:                company.CUI,
		RegistrationNumber: registrationNumber,
		Iban:               iban,
		BankName:           bankName,
		LegalAddress:       legalAddress,
		ContactEmail:       contactEmail,
		ContactPhone:       contactPhone,
		ApprovalStatus:     model.CompanyApprovalStatus(company.ApprovalStatus),
		RejectedReason:     rejectedReason,
		IsActive:           company.IsActive,
		CreatedAt:          company.CreatedAt,
		UpdatedAt:          company.UpdatedAt,
	}
}

// convertCompanyCleanerToGraphQL converts database company cleaner model to GraphQL model
func convertCompanyCleanerToGraphQL(companyCleaner *models.CompanyCleaner, r *Resolver) *model.CompanyCleaner {
	var leftAt *time.Time

	if companyCleaner.LeftAt.Valid {
		leftAt = &companyCleaner.LeftAt.Time
	}

	// Get the cleaner details
	cleaner, _ := r.CleanerService.GetCleanerByID(companyCleaner.CleanerID)
	var cleanerGraphQL *model.Cleaner
	if cleaner != nil {
		cleanerGraphQL = convertCleanerToGraphQL(cleaner)
	}

	return &model.CompanyCleaner{
		ID:        companyCleaner.ID,
		CompanyID: companyCleaner.CompanyID,
		CleanerID: companyCleaner.CleanerID,
		Cleaner:   cleanerGraphQL,
		Status:    companyCleaner.Status,
		JoinedAt:  companyCleaner.JoinedAt,
		LeftAt:    leftAt,
	}
}

// convertCheckinToGraphQL converts database checkin model to GraphQL model
func convertCheckinToGraphQL(checkin *models.Checkin) *model.Checkin {
	var checkInTime, checkOutTime *time.Time
	var checkInLat, checkInLng, checkOutLat, checkOutLng, totalHours *float64

	if checkin.CheckInTime.Valid {
		checkInTime = &checkin.CheckInTime.Time
	}
	if checkin.CheckInLatitude.Valid {
		checkInLat = &checkin.CheckInLatitude.Float64
	}
	if checkin.CheckInLongitude.Valid {
		checkInLng = &checkin.CheckInLongitude.Float64
	}
	if checkin.CheckOutTime.Valid {
		checkOutTime = &checkin.CheckOutTime.Time
	}
	if checkin.CheckOutLatitude.Valid {
		checkOutLat = &checkin.CheckOutLatitude.Float64
	}
	if checkin.CheckOutLongitude.Valid {
		checkOutLng = &checkin.CheckOutLongitude.Float64
	}
	if checkin.TotalHoursWorked.Valid {
		totalHours = &checkin.TotalHoursWorked.Float64
	}

	return &model.Checkin{
		ID:                checkin.ID,
		BookingID:         checkin.BookingID,
		CleanerID:         checkin.CleanerID,
		CheckInTime:       checkInTime,
		CheckInLatitude:   checkInLat,
		CheckInLongitude:  checkInLng,
		CheckOutTime:      checkOutTime,
		CheckOutLatitude:  checkOutLat,
		CheckOutLongitude: checkOutLng,
		TotalHoursWorked:  totalHours,
		CreatedAt:         checkin.CreatedAt,
		UpdatedAt:         checkin.UpdatedAt,
	}
}

// convertInvoiceToGraphQL converts database invoice model to GraphQL model
func convertInvoiceToGraphQL(invoice *models.Invoice) *model.Invoice {
	var clientEmail, pdfURL, xmlURL *string
	var anafUploadIndex, anafDownloadID, anafConfirmationURL *string
	var anafSubmittedAt, anafProcessedAt, anafLastRetryAt *time.Time

	if invoice.ClientEmail.Valid {
		clientEmail = &invoice.ClientEmail.String
	}
	if invoice.PdfURL.Valid {
		pdfURL = &invoice.PdfURL.String
	}
	if invoice.XmlURL.Valid {
		xmlURL = &invoice.XmlURL.String
	}

	// ANAF fields
	if invoice.ANAFUploadIndex.Valid {
		anafUploadIndex = &invoice.ANAFUploadIndex.String
	}
	if invoice.ANAFDownloadID.Valid {
		anafDownloadID = &invoice.ANAFDownloadID.String
	}
	if invoice.ANAFConfirmationURL.Valid {
		anafConfirmationURL = &invoice.ANAFConfirmationURL.String
	}
	if invoice.ANAFSubmittedAt.Valid {
		anafSubmittedAt = &invoice.ANAFSubmittedAt.Time
	}
	if invoice.ANAFProcessedAt.Valid {
		anafProcessedAt = &invoice.ANAFProcessedAt.Time
	}
	if invoice.ANAFLastRetryAt.Valid {
		anafLastRetryAt = &invoice.ANAFLastRetryAt.Time
	}

	// Convert ANAF errors
	var anafErrors []*model.ANAFError
	if len(invoice.ANAFErrors) > 0 {
		anafErrors = make([]*model.ANAFError, len(invoice.ANAFErrors))
		for i, err := range invoice.ANAFErrors {
			fieldPtr := &err.Field
			if err.Field == "" {
				fieldPtr = nil
			}
			anafErrors[i] = &model.ANAFError{
				Code:    err.Code,
				Message: err.Message,
				Field:   fieldPtr,
			}
		}
	}

	return &model.Invoice{
		ID:                 invoice.ID,
		BookingID:          invoice.BookingID,
		InvoiceNumber:      invoice.InvoiceNumber,
		IssueDate:          invoice.IssueDate,
		DueDate:            invoice.DueDate,
		ClientName:         invoice.ClientName,
		ClientEmail:        clientEmail,
		CleanerName:        invoice.CleanerName,
		ServiceDescription: invoice.ServiceDescription,
		Subtotal:           invoice.Subtotal,
		TaxAmount:          invoice.TaxAmount,
		TotalAmount:        invoice.TotalAmount,
		Currency:           invoice.Currency,
		Status:             model.InvoiceStatus(invoice.Status),
		PDFURL:             pdfURL,
		XMLURL:             xmlURL,
		CreatedAt:          invoice.CreatedAt,
		UpdatedAt:          invoice.UpdatedAt,

		// ANAF e-Factura fields
		AnafUploadIndex:     anafUploadIndex,
		AnafStatus:          model.ANAFStatus(invoice.ANAFStatus),
		AnafSubmittedAt:     anafSubmittedAt,
		AnafProcessedAt:     anafProcessedAt,
		AnafDownloadID:      anafDownloadID,
		AnafConfirmationURL: anafConfirmationURL,
		AnafErrors:          anafErrors,
		AnafRetryCount:      invoice.ANAFRetryCount,
		AnafLastRetryAt:     anafLastRetryAt,
	}
}

// convertReviewToGraphQL converts database review model to GraphQL model
func convertReviewToGraphQL(review *models.Review) *model.Review {
	var comment *string

	if review.Comment.Valid {
		comment = &review.Comment.String
	}

	return &model.Review{
		ID:           review.ID,
		BookingID:    review.BookingID,
		ReviewerID:   review.ReviewerID,
		RevieweeID:   review.RevieweeID,
		ReviewerRole: model.ReviewerRole(review.ReviewerRole),
		Rating:       review.Rating,
		Comment:      comment,
		CreatedAt:    review.CreatedAt,
		UpdatedAt:    review.UpdatedAt,
	}
}

// convertDisputeToGraphQL converts database dispute model to GraphQL model
func convertDisputeToGraphQL(dispute *models.Dispute) *model.Dispute{
	var assignedTo, resolutionNotes, resolvedBy, cleanerResponse *string
	var resolvedAt, cleanerRespondedAt *time.Time
	var refundAmount *float64

	if dispute.AssignedTo.Valid {
		assignedTo = &dispute.AssignedTo.String
	}
	if dispute.ResolutionNotes.Valid {
		resolutionNotes = &dispute.ResolutionNotes.String
	}
	if dispute.RefundAmount.Valid {
		refundAmount = &dispute.RefundAmount.Float64
	}
	if dispute.ResolvedAt.Valid {
		resolvedAt = &dispute.ResolvedAt.Time
	}
	if dispute.ResolvedBy.Valid {
		resolvedBy = &dispute.ResolvedBy.String
	}
	if dispute.CleanerResponse.Valid {
		cleanerResponse = &dispute.CleanerResponse.String
	}
	if dispute.CleanerRespondedAt.Valid {
		cleanerRespondedAt = &dispute.CleanerRespondedAt.Time
	}

	// Convert resolution type properly
	var resType *model.DisputeResolutionType
	if dispute.ResolutionType.Valid {
		rt := model.DisputeResolutionType(dispute.ResolutionType.String)
		resType = &rt
	}

	return &model.Dispute{
		ID:                     dispute.ID,
		BookingID:              dispute.BookingID,
		CreatedBy:              dispute.CreatedBy,
		AssignedTo:             assignedTo,
		DisputeType:            model.DisputeType(dispute.DisputeType),
		Status:                 model.DisputeStatus(dispute.Status),
		Description:            dispute.Description,
		ResolutionType:         resType,
		ResolutionNotes:        resolutionNotes,
		RefundAmount:           refundAmount,
		ResolvedAt:             resolvedAt,
		ResolvedBy:             resolvedBy,
		CleanerResponse:        cleanerResponse,
		CleanerRespondedAt:     cleanerRespondedAt,
		CreatedAt:              dispute.CreatedAt,
		UpdatedAt:              dispute.UpdatedAt,
	}
}

// convertPhotoToGraphQL converts database photo model to GraphQL model
func convertPhotoToGraphQL(photo *models.Photo) *model.Photo {
	var bookingID, disputeID *string

	if photo.BookingID != "" {
		bookingID = &photo.BookingID
	}
	if photo.DisputeID.Valid {
		disputeID = &photo.DisputeID.String
	}

	return &model.Photo{
		ID:         photo.ID,
		BookingID:  bookingID,
		DisputeID:  disputeID,
		UploadedBy: photo.UploadedBy,
		PhotoType:  model.PhotoType(photo.PhotoType),
		FileName:   photo.FileName,
		FileSize:   photo.FileSize,
		MimeType:   photo.MimeType,
		URL:        photo.URL,
		CreatedAt:  photo.CreatedAt,
		UpdatedAt:  photo.UpdatedAt,
	}
}

// convertPayoutToGraphQL converts database payout model to GraphQL model
func convertPayoutToGraphQL(payout *models.Payout) *model.Payout {
	var iban, transferRef, invoiceURL, failedReason *string
	var paidAt *time.Time

	if payout.IBAN.Valid {
		iban = &payout.IBAN.String
	}
	if payout.TransferReference.Valid {
		transferRef = &payout.TransferReference.String
	}
	if payout.SettlementInvoiceURL.Valid {
		invoiceURL = &payout.SettlementInvoiceURL.String
	}
	if payout.PaidAt.Valid {
		paidAt = &payout.PaidAt.Time
	}
	if payout.FailedReason.Valid {
		failedReason = &payout.FailedReason.String
	}

	return &model.Payout{
		ID:                   payout.ID,
		CleanerID:            payout.CleanerID,
		PeriodStart:          payout.PeriodStart,
		PeriodEnd:            payout.PeriodEnd,
		Status:               model.PayoutStatus(payout.Status),
		TotalBookings:        payout.TotalBookings,
		TotalEarnings:        payout.TotalEarnings,
		PlatformFees:         payout.PlatformFees,
		NetAmount:            payout.NetAmount,
		Iban:                 iban,
		TransferReference:    transferRef,
		SettlementInvoiceURL: invoiceURL,
		PaidAt:               paidAt,
		FailedReason:         failedReason,
		CreatedAt:            payout.CreatedAt,
		UpdatedAt:            payout.UpdatedAt,
		LineItems:            []*model.PayoutLineItem{}, // Will be loaded via field resolver
	}
}

// convertPayoutToGraphQLWithLineItems converts payout with line items
func convertPayoutToGraphQLWithLineItems(payout *models.Payout, lineItems []*models.PayoutLineItem) *model.Payout {
	result := convertPayoutToGraphQL(payout)
	
	result.LineItems = make([]*model.PayoutLineItem, len(lineItems))
	for i, item := range lineItems {
		result.LineItems[i] = convertPayoutLineItemToGraphQL(item)
	}
	
	return result
}

// convertPayoutLineItemToGraphQL converts line item to GraphQL model
func convertPayoutLineItemToGraphQL(item *models.PayoutLineItem) *model.PayoutLineItem {
	return &model.PayoutLineItem{
		ID:              item.ID,
		PayoutID:        item.PayoutID,
		BookingID:       item.BookingID,
		BookingDate:     item.BookingDate,
		ServiceType:     item.ServiceType,
		BookingAmount:   item.BookingAmount,
		PlatformFeeRate: item.PlatformFeeRate,
		PlatformFee:     item.PlatformFee,
		CleanerEarnings: item.CleanerEarnings,
		CreatedAt:       item.CreatedAt,
	}
}

// convertPlatformSettingsToGraphQL converts database platform settings to GraphQL model
func convertPlatformSettingsToGraphQL(settings *models.PlatformSettings) *model.PlatformSettings {
	return &model.PlatformSettings{
		ID:                        settings.ID.String(),
		BasePrice:                 settings.BasePrice,
		WeekendMultiplier:         settings.WeekendMultiplier,
		EveningMultiplier:         settings.EveningMultiplier,
		PlatformFeePercent:        settings.PlatformFeePercent,
		EmailNotificationsEnabled: settings.EmailNotificationsEnabled,
		AutoApprovalEnabled:       settings.AutoApprovalEnabled,
		MaintenanceMode:           settings.MaintenanceMode,
		UpdatedAt:                 settings.UpdatedAt,
	}
}

// convertMessageToGraphQL converts database message model to GraphQL model
func convertMessageToGraphQL(message *models.Message) *model.Message {
	return &model.Message{
		ID:         message.ID,
		BookingID:  message.BookingID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		IsRead:     message.IsRead,
		CreatedAt:  message.CreatedAt,
		UpdatedAt:  message.UpdatedAt,
	}
}

// convertCleanerStatsToGraphQL converts database cleaner stats model to GraphQL model
func convertCleanerStatsToGraphQL(stats *models.CleanerStats) *model.CleanerStats {
	var averageRating *float64
	var responseTime *float64
	var lastActiveDate *time.Time

	if stats.AverageRating.Valid {
		averageRating = &stats.AverageRating.Float64
	}
	if stats.ResponseTime.Valid {
		responseTime = &stats.ResponseTime.Float64
	}
	if stats.LastActiveDate.Valid {
		lastActiveDate = &stats.LastActiveDate.Time
	}

	return &model.CleanerStats{
		TotalBookings:     stats.TotalBookings,
		CompletedBookings: stats.CompletedBookings,
		CancelledBookings: stats.CancelledBookings,
		NoShowCount:       stats.NoShowCount,
		AverageRating:     averageRating,
		TotalEarnings:     stats.TotalEarnings,
		CompletionRate:    stats.CompletionRate,
		ResponseTime:      responseTime,
		LastActiveDate:    lastActiveDate,
	}
}
