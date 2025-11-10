package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Redis        RedisConfig        `yaml:"redis"`
	Auth         AuthConfig         `yaml:"auth"`
	Pricing      PricingConfig      `yaml:"pricing"`
	Booking      BookingConfig      `yaml:"booking"`
	Cleaner      CleanerConfig      `yaml:"cleaner"`
	Client       ClientConfig       `yaml:"client"`
	Company      CompanyConfig      `yaml:"company"`
	ANAF         ANAFConfig         `yaml:"anaf"`
	Payment      PaymentConfig      `yaml:"payment"`
	Notification NotificationConfig `yaml:"notifications"`
	Features     FeaturesConfig     `yaml:"features"`
	Business     BusinessConfig     `yaml:"business"`
	EFactura     EFacturaConfig     `yaml:"efactura"` // Legacy - use ANAF instead
}

type ServerConfig struct {
	Port        int    `yaml:"port"`
	Environment string `yaml:"environment"`
}

type DatabaseConfig struct {
	MaxOpenConnections        int `yaml:"max_open_connections"`
	MaxIdleConnections        int `yaml:"max_idle_connections"`
	ConnectionMaxLifetimeMin  int `yaml:"connection_max_lifetime_minutes"`
}

type RedisConfig struct {
	CacheTTLMinutes int `yaml:"cache_ttl_minutes"`
}

type AuthConfig struct {
	JWTExpirationDays     int    `yaml:"jwt_expiration_days"`
	OTPExpirationMinutes  int    `yaml:"otp_expiration_minutes"`
	OTPLength             int    `yaml:"otp_length"`
	OTPDevCode            string `yaml:"otp_dev_code"`
}

type PricingConfig struct {
	DefaultPlatformFeePercentage      float64             `yaml:"default_platform_fee_percentage"`
	FirstBookingDiscountPercentage    float64             `yaml:"first_booking_discount_percentage"`
	RepeatCustomerDiscountPercentage  float64             `yaml:"repeat_customer_discount_percentage"`
	StandardCleaning                  ServicePricing      `yaml:"standard_cleaning"`
	DeepCleaning                      DeepCleaningPricing `yaml:"deep_cleaning"`
	OfficeCleaning                    ServicePricing      `yaml:"office_cleaning"`
	PostRenovation                    ServicePricing      `yaml:"post_renovation"`
	MoveInOut                         ServicePricing      `yaml:"move_in_out"`
	Addons                            AddonsPricing       `yaml:"addons"`
	Multipliers                       Multipliers         `yaml:"multipliers"`
	FrequencyDiscounts                FrequencyDiscounts  `yaml:"frequency_discounts"`
}

type FrequencyDiscounts struct {
	Weekly    float64 `yaml:"weekly"`     // Percentage discount for weekly bookings
	Biweekly  float64 `yaml:"biweekly"`   // Percentage discount for biweekly bookings
	Monthly   float64 `yaml:"monthly"`    // Percentage discount for monthly bookings
}

type ServicePricing struct {
	BasePricePerHour float64 `yaml:"base_price_per_hour"`
	MinimumHours     int     `yaml:"minimum_hours"`
	PricePerSqm      float64 `yaml:"price_per_sqm"`
}

type DeepCleaningPricing struct {
	ServicePricing `yaml:",inline"`
	Multiplier     float64 `yaml:"multiplier"`
}

type AddonsPricing struct {
	WindowCleaningPerWindow float64 `yaml:"window_cleaning_per_window"`
	CarpetCleaningPerSqm    float64 `yaml:"carpet_cleaning_per_sqm"`
	FridgeCleaning          float64 `yaml:"fridge_cleaning"`
	OvenCleaning            float64 `yaml:"oven_cleaning"`
	BalconyCleaning         float64 `yaml:"balcony_cleaning"`
	CleaningSupplies        float64 `yaml:"cleaning_supplies"`
}

type Multipliers struct {
	Weekend float64 `yaml:"weekend"`
	Evening float64 `yaml:"evening"`
	Holiday float64 `yaml:"holiday"`
}

type BookingConfig struct {
	MinAdvanceBookingHours   int `yaml:"min_advance_booking_hours"`
	MaxAdvanceBookingDays    int `yaml:"max_advance_booking_days"`
	CancellationFreeHours    int `yaml:"cancellation_free_hours"`
	CleanerSearchRadiusKm    int `yaml:"cleaner_search_radius_km"`
	AutoAssignTimeoutMinutes int `yaml:"auto_assign_timeout_minutes"`
	MinRating                int `yaml:"min_rating"`
	MaxRating                int `yaml:"max_rating"`
}

type CleanerConfig struct {
	RequireIDDocument         bool    `yaml:"require_id_document"`
	RequireBackgroundCheck    bool    `yaml:"require_background_check"`
	RequireProfilePhoto       bool    `yaml:"require_profile_photo"`
	AutoApproveEnabled        bool    `yaml:"auto_approve_enabled"`
	ApprovalTimeoutDays       int     `yaml:"approval_timeout_days"`
	MinRatingToStayActive     float64 `yaml:"min_rating_to_stay_active"`
	MaxCancellationsPerMonth  int     `yaml:"max_cancellations_per_month"`
	MaxNoShowsPerMonth        int     `yaml:"max_no_shows_per_month"`
}

type ClientConfig struct {
	MaxActiveBookings int `yaml:"max_active_bookings"`
	MaxAddresses      int `yaml:"max_addresses"`
}

type PaymentConfig struct {
	Provider             string `yaml:"provider"`
	PreauthEnabled       bool   `yaml:"preauth_enabled"`
	CaptureOnCompletion  bool   `yaml:"capture_on_completion"`
	RefundWindowDays     int    `yaml:"refund_window_days"`
}

type NotificationConfig struct {
	EmailEnabled    bool   `yaml:"email_enabled"`
	SMSEnabled      bool   `yaml:"sms_enabled"`
	PushEnabled     bool   `yaml:"push_enabled"`
	EmailFrom       string `yaml:"email_from"`
	EmailFromName   string `yaml:"email_from_name"`
}

type FeaturesConfig struct {
	CleaningCompaniesEnabled bool `yaml:"cleaning_companies_enabled"`
	InstantBookingEnabled    bool `yaml:"instant_booking_enabled"`
	RecurringBookingsEnabled bool `yaml:"recurring_bookings_enabled"`
	GiftCardsEnabled         bool `yaml:"gift_cards_enabled"`
}

type BusinessConfig struct {
	ServiceStartHour     int      `yaml:"service_start_hour"`
	ServiceEndHour       int      `yaml:"service_end_hour"`
	EnabledCities        []string `yaml:"enabled_cities"`
	SupportedLanguages   []string `yaml:"supported_languages"`
	DefaultLanguage      string   `yaml:"default_language"`
}

type EFacturaConfig struct {
	Enabled            bool   `yaml:"enabled"`
	Environment        string `yaml:"environment"`
	RetryAttempts      int    `yaml:"retry_attempts"`
	RetryDelaySeconds  int    `yaml:"retry_delay_seconds"`
}

// CompanyConfig holds company information for invoices
type CompanyConfig struct {
	LegalName          string        `yaml:"legal_name"`
	TradeName          string        `yaml:"trade_name"`
	CUI                string        `yaml:"cui"`
	RegistrationNumber string        `yaml:"registration_number"`
	VATRegistered      bool          `yaml:"vat_registered"`
	VATRate            float64       `yaml:"vat_rate"`
	Address            CompanyAddress `yaml:"address"`
	Contact            CompanyContact `yaml:"contact"`
	Bank               CompanyBank    `yaml:"bank"`
}

type CompanyAddress struct {
	Street     string `yaml:"street"`
	City       string `yaml:"city"`
	County     string `yaml:"county"`
	PostalCode string `yaml:"postal_code"`
	Country    string `yaml:"country"`
}

type CompanyContact struct {
	Email   string `yaml:"email"`
	Phone   string `yaml:"phone"`
	Website string `yaml:"website"`
}

type CompanyBank struct {
	Name  string `yaml:"name"`
	IBAN  string `yaml:"iban"`
	SWIFT string `yaml:"swift"`
}

// ANAFConfig holds ANAF e-Factura configuration
type ANAFConfig struct {
	Enabled                bool   `yaml:"enabled"`
	Environment            string `yaml:"environment"`
	SandboxURL             string `yaml:"sandbox_url"`
	ProductionURL          string `yaml:"production_url"`
	ClientID               string `yaml:"client_id"`
	ClientSecret           string `yaml:"client_secret"`
	AutoSubmit             bool   `yaml:"auto_submit"`
	SubmissionDelayMinutes int    `yaml:"submission_delay_minutes"`
	MaxRetryAttempts       int    `yaml:"max_retry_attempts"`
	RetryDelayMinutes      int    `yaml:"retry_delay_minutes"`
	SubmissionDeadlineDays int    `yaml:"submission_deadline_days"`
}

var appConfig *Config

// Load loads configuration from YAML file
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try from project root
		rootPath := filepath.Join("..", configPath)
		if _, err := os.Stat(rootPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", configPath)
		}
		configPath = rootPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	appConfig = &config
	return &config, nil
}

// Get returns the global config instance
func Get() *Config {
	if appConfig == nil {
		panic("config not loaded - call Load() first")
	}
	return appConfig
}

// MustLoad loads config or panics
func MustLoad(configPath string) *Config {
	config, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}
