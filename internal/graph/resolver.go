package graph

import (
	"github.com/cleanbuddy/backend/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	AuthService                  *services.AuthService
	ClientService                *services.ClientService
	AddressService               *services.AddressService
	CleanerService               *services.CleanerService
	BookingService               *services.BookingService
	PricingService               *services.PricingService
	PaymentService               *services.PaymentService
	AvailabilityService          *services.AvailabilityService
	CompanyService               *services.CompanyService
	CheckinService               *services.CheckinService
	InvoiceService               *services.InvoiceService
	ReviewService                *services.ReviewService
	DisputeService               *services.DisputeService
	PhotoService                 *services.PhotoService
	PayoutService                *services.PayoutService
	AdminAnalyticsService        *services.AdminAnalyticsService
	PlatformSettingsService      *services.PlatformSettingsService
	MessagingService             *services.MessagingService
	CleanerApplicationService    *services.CleanerApplicationService
}
