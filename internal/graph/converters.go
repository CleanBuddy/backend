package graph

import (
	"github.com/cleanbuddy/backend/internal/graph/model"
	"github.com/cleanbuddy/backend/internal/models"
)

// convertCleanerApplicationToGraphQL converts cleaner application to GraphQL model
func convertCleanerApplicationToGraphQL(app *models.CleanerApplication) *model.CleanerApplication {
	if app == nil {
		return nil
	}

	gqlApp := &model.CleanerApplication{
		ID:              app.ID.String(),
		CurrentStep:     app.CurrentStep,
		Status:          model.ApplicationStatus(app.Status),
		ApplicationData: convertCleanerApplicationDataToGraphQL(&app.ApplicationData),
		CreatedAt:       app.CreatedAt,
		UpdatedAt:       app.UpdatedAt,
	}

	if app.SessionID != nil {
		gqlApp.SessionID = app.SessionID
	}

	if app.UserID != nil {
		userID := app.UserID.String()
		gqlApp.UserID = &userID
	}

	if app.ReviewedBy != nil {
		// TODO: Load reviewer user object
		// For now, just skip this field
	}

	if app.ReviewedAt != nil {
		gqlApp.ReviewedAt = app.ReviewedAt
	}

	if app.RejectionReason != nil {
		gqlApp.RejectionReason = app.RejectionReason
	}

	if app.AdminNotes != nil {
		gqlApp.AdminNotes = app.AdminNotes
	}

	if app.ConvertedToCleanerID != nil {
		cleanerID := app.ConvertedToCleanerID.String()
		gqlApp.ConvertedToCleanerID = &cleanerID
	}

	if app.SubmittedAt != nil {
		gqlApp.SubmittedAt = app.SubmittedAt
	}

	return gqlApp
}

// convertCleanerApplicationDataToGraphQL converts application data to GraphQL model
func convertCleanerApplicationDataToGraphQL(data *models.CleanerApplicationData) *model.CleanerApplicationData {
	gqlData := &model.CleanerApplicationData{}

	if data.Eligibility != nil {
		gqlData.Eligibility = &model.EligibilityData{
			Age18Plus:  data.Eligibility.Age18Plus,
			WorkRight:  data.Eligibility.WorkRight,
			Experience: data.Eligibility.Experience,
		}
	}

	if data.Availability != nil {
		gqlData.Availability = &model.AvailabilityData{
			HoursPerWeek:             data.Availability.HoursPerWeek,
			Areas:                    data.Availability.Areas,
			Days:                     data.Availability.Days,
			TimeSlots:                data.Availability.TimeSlots,
			EstimatedMonthlyEarnings: data.Availability.EstimatedMonthlyEarnings,
		}
	}

	if data.Profile != nil {
		gqlData.Profile = &model.ProfileData{
			PhotoURL:  data.Profile.PhotoURL,
			Bio:       data.Profile.Bio,
			Languages: data.Profile.Languages,
			Equipment: data.Profile.Equipment,
		}
	}

	if data.Legal != nil {
		gqlData.Legal = &model.LegalData{
			Status:       data.Legal.Status,
			Cif:          data.Legal.CIF,
			CnpEncrypted: data.Legal.CNPEncrypted,
			Iban:         data.Legal.IBAN,
			BankName:     data.Legal.BankName,
		}
	}

	if data.Documents != nil {
		gqlData.Documents = &model.DocumentData{
			CazierURL:    data.Documents.CazierURL,
			IDFrontURL:   data.Documents.IDFrontURL,
			IDBackURL:    data.Documents.IDBackURL,
			InsuranceURL: data.Documents.InsuranceURL,
		}
	}

	return gqlData
}
