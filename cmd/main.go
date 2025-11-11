package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/db"
	"github.com/cleanbuddy/backend/internal/graph"
	"github.com/cleanbuddy/backend/internal/graph/generated"
	"github.com/cleanbuddy/backend/internal/middleware"
	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("✅ Loaded configuration from %s", configPath)

	// Initialize database connection
	database, err := db.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize Redis client (optional)
	var redisClient *redis.Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("⚠️  Failed to parse Redis URL: %v (falling back to in-memory storage)", err)
	} else {
		redisClient = redis.NewClient(opt)

		// Test Redis connection (non-fatal)
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("⚠️  Failed to connect to Redis: %v (falling back to in-memory storage)", err)
			redisClient.Close()
			redisClient = nil
		} else {
			defer redisClient.Close()
			log.Println("✅ Connected to Redis")
		}
	}

	// Initialize services
	emailService := services.NewEmailService()
	authService := services.NewAuthService(database.DB, redisClient, emailService)
	clientService := services.NewClientService(database.DB)
	addressService := services.NewAddressService(database.DB)
	cleanerService := services.NewCleanerService(database.DB, emailService)
	pricingService := services.NewPricingService(database.DB)
	invoiceService := services.NewInvoiceService(database.DB, &cfg.Company, &cfg.ANAF)
	reviewService := services.NewReviewService(database.DB)
	disputeService := services.NewDisputeService(database.DB)
	photoService := services.NewPhotoService(database.DB, "./uploads")
	payoutService := services.NewPayoutService(database.DB, emailService)
	bookingService := services.NewBookingService(database.DB, pricingService, invoiceService, emailService)
	paymentService := services.NewPaymentService(database.DB)
	matchingService := services.NewCleanerMatchingService(database.DB, emailService)
	bookingService.SetPaymentService(paymentService)   // Set payment service after creation
	bookingService.SetMatchingService(matchingService) // Set matching service after creation
	disputeService.SetPaymentService(paymentService)   // Set payment service for refunds
	disputeService.SetBookingService(bookingService)   // Set booking service for recleans
	disputeService.SetEmailService(emailService)       // Set email service for notifications
	availabilityService := services.NewAvailabilityService(database.DB)
	companyService := services.NewCompanyService(database.DB)
	checkinService := services.NewCheckinService(database.DB, bookingService)
	adminAnalyticsService := services.NewAdminAnalyticsService(database.DB)
	platformSettingsService := services.NewPlatformSettingsService(models.NewPlatformSettingsRepository(database.DB))
	messagingService := services.NewMessagingService(database.DB)
	cleanerApplicationService := services.NewCleanerApplicationService(database.DB)

	// Initialize GraphQL resolver with dependencies
	resolver := &graph.Resolver{
		AuthService:               authService,
		ClientService:             clientService,
		AddressService:            addressService,
		CleanerService:            cleanerService,
		BookingService:            bookingService,
		PricingService:            pricingService,
		PaymentService:            paymentService,
		AvailabilityService:       availabilityService,
		CompanyService:            companyService,
		CheckinService:            checkinService,
		InvoiceService:            invoiceService,
		ReviewService:             reviewService,
		DisputeService:            disputeService,
		PhotoService:              photoService,
		PayoutService:             payoutService,
		AdminAnalyticsService:     adminAnalyticsService,
		PlatformSettingsService:   platformSettingsService,
		MessagingService:          messagingService,
		CleanerApplicationService: cleanerApplicationService,
	}

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow credentials for httpOnly cookies
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3003" // Default for development
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(authService)

	// Response writer middleware (allows resolvers to set cookies)
	responseWriterMiddleware := middleware.ResponseWriterMiddleware()

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(redisClient)
	rateLimitMiddleware := rateLimiter.RateLimitMiddleware()

	// Security headers middleware
	securityHeadersMiddleware := middleware.SecurityHeadersMiddleware()

	// Start rate limiter fallback cache cleanup
	go rateLimiter.CleanupFallbackCache(10 * time.Minute)

	// Setup routes
	http.Handle("/", securityHeadersMiddleware(corsMiddleware(playground.Handler("GraphQL playground", "/graphql"))))
	http.Handle("/graphql", securityHeadersMiddleware(corsMiddleware(rateLimitMiddleware(authMiddleware(responseWriterMiddleware(srv))))))

	// Serve uploaded files
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", securityHeadersMiddleware(corsMiddleware(http.StripPrefix("/uploads/", fs))))

	// Serve invoice files (PDF and XML)
	invoiceFS := http.FileServer(http.Dir("./invoices"))
	http.Handle("/invoices/", securityHeadersMiddleware(corsMiddleware(http.StripPrefix("/invoices/", invoiceFS))))

	// Start booking expiration scheduler (runs every hour)
	startBookingExpirationScheduler(bookingService)

	log.Printf("🚀 CleanBuddy API server ready at http://localhost:%s/", port)
	log.Printf("📊 GraphQL playground at http://localhost:%s/", port)
	log.Printf("⏰ Booking expiration scheduler running (checks every hour)")
	log.Printf("🛡️  Rate limiting active (Anonymous: 20/min, Authenticated: 100/min)")
	log.Printf("🔒 Security headers enabled (CSP, HSTS, X-Frame-Options, etc.)")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// startBookingExpirationScheduler runs a background task to auto-expire old bookings
func startBookingExpirationScheduler(bookingService *services.BookingService) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		expireBookings(bookingService)

		// Then run every hour
		for range ticker.C {
			expireBookings(bookingService)
		}
	}()
}

func expireBookings(bookingService *services.BookingService) {
	const expirationHours = 24
	count, err := bookingService.ExpireOldPendingBookings(expirationHours)
	if err != nil {
		log.Printf("❌ Error expiring bookings: %v", err)
		return
	}
	if count > 0 {
		log.Printf("✅ Expired %d pending bookings older than %d hours", count, expirationHours)
	}
}
