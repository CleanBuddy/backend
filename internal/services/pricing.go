package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// PriceQuote represents a price calculation result
type PriceQuote struct {
	BasePrice       float64
	AddonsPrice     float64
	Subtotal        float64
	Discount        float64
	PlatformFee     float64
	TotalPrice      float64
	CleanerPayout   float64
	EstimatedHours  int
	Breakdown       PriceBreakdown
}

// PriceBreakdown shows detailed price calculation
type PriceBreakdown struct {
	BasePricePerHour     float64
	HoursCharged         int
	AreaPrice            float64
	WindowsPrice         float64
	CarpetPrice          float64
	TimeMultiplier       float64
	DiscountPercentage   float64
	PlatformFeePercentage float64
}

// PricingService handles pricing calculations
type PricingService struct {
	clientRepo *models.ClientRepository
	cfg        *config.Config
}

// NewPricingService creates a new pricing service
func NewPricingService(db *sql.DB) *PricingService {
	return &PricingService{
		clientRepo: models.NewClientRepository(db),
		cfg:        config.Get(),
	}
}

// CalculatePrice calculates the total price for a booking
func (s *PricingService) CalculatePrice(
	clientID string,
	serviceType models.ServiceType,
	areaSqm int,
	estimatedHours int,
	scheduledDate time.Time,
	scheduledTime time.Time,
	includesWindows bool,
	numberOfWindows int,
	includesCarpet bool,
	carpetAreaSqm int,
	includesFridge bool,
	includesOven bool,
	includesBalcony bool,
	includesSupplies bool,
	frequency string,
) (*PriceQuote, error) {
	// Get service pricing from config
	servicePricing := s.getServicePricing(serviceType)
	if servicePricing == nil {
		return nil, fmt.Errorf("invalid service type: %s", serviceType)
	}

	// Calculate base price
	hoursToCharge := estimatedHours
	if hoursToCharge < servicePricing.MinimumHours {
		hoursToCharge = servicePricing.MinimumHours
	}

	basePrice := servicePricing.BasePricePerHour * float64(hoursToCharge)

	// Add area-based pricing if applicable
	areaPrice := 0.0
	if areaSqm > 0 && servicePricing.PricePerSqm > 0 {
		areaPrice = servicePricing.PricePerSqm * float64(areaSqm)
	}

	// Calculate add-ons
	addonsPrice := 0.0
	windowsPrice := 0.0
	if includesWindows && numberOfWindows > 0 {
		windowsPrice = s.cfg.Pricing.Addons.WindowCleaningPerWindow * float64(numberOfWindows)
		addonsPrice += windowsPrice
	}

	carpetPrice := 0.0
	if includesCarpet && carpetAreaSqm > 0 {
		carpetPrice = s.cfg.Pricing.Addons.CarpetCleaningPerSqm * float64(carpetAreaSqm)
		addonsPrice += carpetPrice
	}

	// Add fixed-price addons
	if includesFridge {
		addonsPrice += s.cfg.Pricing.Addons.FridgeCleaning
	}
	if includesOven {
		addonsPrice += s.cfg.Pricing.Addons.OvenCleaning
	}
	if includesBalcony {
		addonsPrice += s.cfg.Pricing.Addons.BalconyCleaning
	}
	if includesSupplies {
		addonsPrice += s.cfg.Pricing.Addons.CleaningSupplies
	}

	// Apply time-based multipliers
	timeMultiplier := s.getTimeMultiplier(scheduledDate, scheduledTime)

	subtotal := (basePrice + areaPrice + addonsPrice) * timeMultiplier

	// Check if first booking for discount
	isFirstBooking, err := s.isFirstBooking(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check booking history: %w", err)
	}

	discount := 0.0
	discountPercentage := 0.0
	if isFirstBooking {
		discountPercentage = s.cfg.Pricing.FirstBookingDiscountPercentage
		discount = subtotal * (discountPercentage / 100.0)
	}

	// Apply frequency discount (stacks with first booking discount)
	frequencyDiscountPercentage := s.getFrequencyDiscount(frequency)
	if frequencyDiscountPercentage > 0 {
		frequencyDiscount := subtotal * (frequencyDiscountPercentage / 100.0)
		discount += frequencyDiscount
		discountPercentage += frequencyDiscountPercentage
	}

	totalAfterDiscount := subtotal - discount

	// Calculate platform fee
	platformFeePercentage := s.cfg.Pricing.DefaultPlatformFeePercentage
	platformFee := totalAfterDiscount * (platformFeePercentage / 100.0)

	// Total price to client
	totalPrice := totalAfterDiscount

	// Cleaner payout (total - platform fee)
	cleanerPayout := totalAfterDiscount - platformFee

	return &PriceQuote{
		BasePrice:      basePrice + areaPrice,
		AddonsPrice:    addonsPrice,
		Subtotal:       subtotal,
		Discount:       discount,
		PlatformFee:    platformFee,
		TotalPrice:     totalPrice,
		CleanerPayout:  cleanerPayout,
		EstimatedHours: hoursToCharge,
		Breakdown: PriceBreakdown{
			BasePricePerHour:      servicePricing.BasePricePerHour,
			HoursCharged:          hoursToCharge,
			AreaPrice:             areaPrice,
			WindowsPrice:          windowsPrice,
			CarpetPrice:           carpetPrice,
			TimeMultiplier:        timeMultiplier,
			DiscountPercentage:    discountPercentage,
			PlatformFeePercentage: platformFeePercentage,
		},
	}, nil
}

// getServicePricing returns pricing config for a service type
func (s *PricingService) getServicePricing(serviceType models.ServiceType) *servicePricingConfig {
	switch serviceType {
	case models.ServiceTypeStandard:
		return &servicePricingConfig{
			BasePricePerHour: s.cfg.Pricing.StandardCleaning.BasePricePerHour,
			MinimumHours:     s.cfg.Pricing.StandardCleaning.MinimumHours,
			PricePerSqm:      s.cfg.Pricing.StandardCleaning.PricePerSqm,
		}
	case models.ServiceTypeDeepCleaning:
		return &servicePricingConfig{
			BasePricePerHour: s.cfg.Pricing.DeepCleaning.BasePricePerHour,
			MinimumHours:     s.cfg.Pricing.DeepCleaning.MinimumHours,
			PricePerSqm:      s.cfg.Pricing.DeepCleaning.PricePerSqm,
		}
	case models.ServiceTypeOffice:
		return &servicePricingConfig{
			BasePricePerHour: s.cfg.Pricing.OfficeCleaning.BasePricePerHour,
			MinimumHours:     s.cfg.Pricing.OfficeCleaning.MinimumHours,
			PricePerSqm:      s.cfg.Pricing.OfficeCleaning.PricePerSqm,
		}
	case models.ServiceTypePostRenovation:
		return &servicePricingConfig{
			BasePricePerHour: s.cfg.Pricing.PostRenovation.BasePricePerHour,
			MinimumHours:     s.cfg.Pricing.PostRenovation.MinimumHours,
			PricePerSqm:      s.cfg.Pricing.PostRenovation.PricePerSqm,
		}
	case models.ServiceTypeMoveInOut:
		return &servicePricingConfig{
			BasePricePerHour: s.cfg.Pricing.MoveInOut.BasePricePerHour,
			MinimumHours:     s.cfg.Pricing.MoveInOut.MinimumHours,
			PricePerSqm:      s.cfg.Pricing.MoveInOut.PricePerSqm,
		}
	default:
		return nil
	}
}

// getTimeMultiplier calculates time-based price multiplier
func (s *PricingService) getTimeMultiplier(scheduledDate time.Time, scheduledTime time.Time) float64 {
	multiplier := 1.0

	// Weekend multiplier (Saturday = 6, Sunday = 0)
	weekday := scheduledDate.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		multiplier *= s.cfg.Pricing.Multipliers.Weekend
	}

	// Evening multiplier (after 18:00)
	hour := scheduledTime.Hour()
	if hour >= 18 {
		multiplier *= s.cfg.Pricing.Multipliers.Evening
	}

	// Holiday multiplier (Romanian public holidays)
	if utils.IsRomanianHoliday(scheduledDate) {
		multiplier *= s.cfg.Pricing.Multipliers.Holiday
	}

	return multiplier
}

// getFrequencyDiscount returns the discount percentage for a given frequency
func (s *PricingService) getFrequencyDiscount(frequency string) float64 {
	switch frequency {
	case "weekly":
		return s.cfg.Pricing.FrequencyDiscounts.Weekly
	case "biweekly":
		return s.cfg.Pricing.FrequencyDiscounts.Biweekly
	case "monthly":
		return s.cfg.Pricing.FrequencyDiscounts.Monthly
	case "one_time", "":
		return 0.0
	default:
		return 0.0
	}
}

// isFirstBooking checks if this is client's first booking
func (s *PricingService) isFirstBooking(clientID string) (bool, error) {
	client, err := s.clientRepo.GetByUserID(clientID)
	if err != nil {
		return false, err
	}
	if client == nil {
		return true, nil
	}

	return client.TotalBookings == 0, nil
}

// servicePricingConfig holds pricing for a service type
type servicePricingConfig struct {
	BasePricePerHour float64
	MinimumHours     int
	PricePerSqm      float64
}
