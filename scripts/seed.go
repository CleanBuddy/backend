//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Seed script for CleanBuddy development database
// Run with: make db-seed
// Or directly: go run scripts/seed.go

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://cleanbuddy:devpassword@localhost:5432/cleanbuddy?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("🌱 Starting database seeding...")

	// Seed data in order (respecting foreign key constraints)
	seedUsers(db)
	seedClients(db)
	seedCleaners(db)
	seedAddresses(db)
	seedAvailability(db)
	seedBookings(db)
	seedPayments(db)
	seedReviews(db)
	seedDisputes(db)
	seedCompanies(db)
	seedPlatformSettings(db)

	log.Println("✅ Database seeding complete!")
	log.Println("")
	log.Println("Test Accounts:")
	log.Println("  Client: client1@test.com")
	log.Println("  Cleaner: cleaner1@test.com, cleaner2@test.com")
	log.Println("  Admin: admin@cleanbuddy.ro")
	log.Println("")
	log.Println("OTP Code (development): 123456")
}

func seedUsers(db *sql.DB) {
	log.Println("👤 Seeding users...")

	users := []map[string]interface{}{
		{
			"id":         "11111111-1111-1111-1111-111111111111",
			"email":      "client1@test.com",
			"first_name": "Maria",
			"last_name":  "Ionescu",
			"role":       "CLIENT",
		},
		{
			"id":         "22222222-2222-2222-2222-222222222222",
			"email":      "client2@test.com",
			"first_name": "Ion",
			"last_name":  "Popescu",
			"role":       "CLIENT",
		},
		{
			"id":         "33333333-3333-3333-3333-333333333333",
			"email":      "cleaner1@test.com",
			"first_name": "Ana",
			"last_name":  "Dumitrescu",
			"role":       "CLEANER",
		},
		{
			"id":         "44444444-4444-4444-4444-444444444444",
			"email":      "cleaner2@test.com",
			"first_name": "Mihai",
			"last_name":  "Georgescu",
			"role":       "CLEANER",
		},
		{
			"id":         "55555555-5555-5555-5555-555555555555",
			"email":      "cleaner3@test.com",
			"first_name": "Elena",
			"last_name":  "Vasilescu",
			"role":       "CLEANER",
		},
		{
			"id":         "99999999-9999-9999-9999-999999999999",
			"email":      "admin@cleanbuddy.ro",
			"first_name": "Admin",
			"last_name":  "CleanBuddy",
			"role":       "PLATFORM_ADMIN",
		},
	}

	for _, user := range users {
		_, err := db.Exec(`
			INSERT INTO users (id, email, first_name, last_name, role, email_verified, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, TRUE, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`, user["id"], user["email"], user["first_name"], user["last_name"], user["role"])

		if err != nil {
			log.Printf("Warning: Failed to seed user %s: %v", user["email"], err)
		}
	}

	log.Printf("   Created %d users", len(users))
}

func seedClients(db *sql.DB) {
	log.Println("👨‍💼 Seeding clients...")

	clients := []map[string]interface{}{
		{
			"id":                "c1111111-1111-1111-1111-111111111111",
			"user_id":           "11111111-1111-1111-1111-111111111111",
			"phone_number":      "+40721234567",
			"total_bookings":    5,
			"total_spent":       650.00,
			"preferred_language": "ro",
		},
		{
			"id":                "c2222222-2222-2222-2222-222222222222",
			"user_id":           "22222222-2222-2222-2222-222222222222",
			"phone_number":      "+40731234567",
			"total_bookings":    2,
			"total_spent":       240.00,
			"preferred_language": "ro",
		},
	}

	for _, client := range clients {
		_, err := db.Exec(`
			INSERT INTO clients (id, user_id, phone_number, total_bookings, total_spent, preferred_language, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`, client["id"], client["user_id"], client["phone_number"], client["total_bookings"], client["total_spent"], client["preferred_language"])

		if err != nil {
			log.Printf("Warning: Failed to seed client: %v", err)
		}
	}

	log.Printf("   Created %d clients", len(clients))
}

func seedCleaners(db *sql.DB) {
	log.Println("🧹 Seeding cleaners...")

	cleaners := []map[string]interface{}{
		{
			"id":                   "cl111111-1111-1111-1111-111111111111",
			"user_id":              "33333333-3333-3333-3333-333333333333",
			"phone_number":         "+40741234567",
			"city":                 "București",
			"county":               "București",
			"years_of_experience":  5,
			"bio":                  "Servicii de curățenie profesionale de 5 ani. Specializare în curățenie profundă.",
			"approval_status":      "APPROVED",
			"tier":                 "STANDARD",
			"average_rating":       4.8,
			"total_jobs":           42,
			"total_earnings":       5040.00,
			"is_active":            true,
		},
		{
			"id":                   "cl222222-2222-2222-2222-222222222222",
			"user_id":              "44444444-4444-4444-4444-444444444444",
			"phone_number":         "+40751234567",
			"city":                 "Cluj-Napoca",
			"county":               "Cluj",
			"years_of_experience":  3,
			"bio":                  "Experiență în curățenie rezidențială și comercială.",
			"approval_status":      "APPROVED",
			"tier":                 "PREMIUM",
			"average_rating":       4.9,
			"total_jobs":           68,
			"total_earnings":       8500.00,
			"is_active":            true,
		},
		{
			"id":                   "cl333333-3333-3333-3333-333333333333",
			"user_id":              "55555555-5555-5555-5555-555555555555",
			"phone_number":         "+40761234567",
			"city":                 "București",
			"county":               "București",
			"years_of_experience":  2,
			"bio":                  "Curățenie eco-friendly cu produse naturale.",
			"approval_status":      "PENDING",
			"tier":                 "STANDARD",
			"average_rating":       nil,
			"total_jobs":           0,
			"total_earnings":       0.00,
			"is_active":            true,
		},
	}

	for _, cleaner := range cleaners {
		query := `
			INSERT INTO cleaners (
				id, user_id, phone_number, city, county, years_of_experience,
				bio, approval_status, tier, average_rating, total_jobs,
				total_earnings, is_active, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`

		_, err := db.Exec(query,
			cleaner["id"], cleaner["user_id"], cleaner["phone_number"],
			cleaner["city"], cleaner["county"], cleaner["years_of_experience"],
			cleaner["bio"], cleaner["approval_status"], cleaner["tier"],
			cleaner["average_rating"], cleaner["total_jobs"],
			cleaner["total_earnings"], cleaner["is_active"])

		if err != nil {
			log.Printf("Warning: Failed to seed cleaner: %v", err)
		}
	}

	log.Printf("   Created %d cleaners", len(cleaners))
}

func seedAddresses(db *sql.DB) {
	log.Println("📍 Seeding addresses...")

	addresses := []map[string]interface{}{
		{
			"id":             "a1111111-1111-1111-1111-111111111111",
			"client_id":      "c1111111-1111-1111-1111-111111111111",
			"label":          "Acasă",
			"street_address": "Strada Republicii 15",
			"city":           "București",
			"county":         "București",
			"postal_code":    "010123",
			"is_default":     true,
		},
		{
			"id":             "a2222222-2222-2222-2222-222222222222",
			"client_id":      "c2222222-2222-2222-2222-222222222222",
			"label":          "Apartament",
			"street_address": "Bulevardul Unirii 45, Ap. 23",
			"apartment":      "23",
			"city":           "Cluj-Napoca",
			"county":         "Cluj",
			"postal_code":    "400000",
			"is_default":     true,
		},
	}

	for _, addr := range addresses {
		query := `
			INSERT INTO addresses (
				id, client_id, label, street_address, apartment, city,
				county, postal_code, is_default, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`

		_, err := db.Exec(query,
			addr["id"], addr["client_id"], addr["label"], addr["street_address"],
			addr["apartment"], addr["city"], addr["county"],
			addr["postal_code"], addr["is_default"])

		if err != nil {
			log.Printf("Warning: Failed to seed address: %v", err)
		}
	}

	log.Printf("   Created %d addresses", len(addresses))
}

func seedAvailability(db *sql.DB) {
	log.Println("📅 Seeding availability...")

	// Create recurring weekly availability for approved cleaners
	availabilities := []map[string]interface{}{
		{
			"cleaner_id":  "cl111111-1111-1111-1111-111111111111",
			"type":        "RECURRING",
			"day_of_week": 1, // Monday
			"start_time":  "09:00",
			"end_time":    "17:00",
			"is_active":   true,
		},
		{
			"cleaner_id":  "cl111111-1111-1111-1111-111111111111",
			"type":        "RECURRING",
			"day_of_week": 3, // Wednesday
			"start_time":  "09:00",
			"end_time":    "17:00",
			"is_active":   true,
		},
		{
			"cleaner_id":  "cl222222-2222-2222-2222-222222222222",
			"type":        "RECURRING",
			"day_of_week": 2, // Tuesday
			"start_time":  "08:00",
			"end_time":    "18:00",
			"is_active":   true,
		},
	}

	for _, avail := range availabilities {
		query := `
			INSERT INTO availability (
				cleaner_id, type, day_of_week, start_time, end_time, is_active, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`

		_, err := db.Exec(query,
			avail["cleaner_id"], avail["type"], avail["day_of_week"],
			avail["start_time"], avail["end_time"], avail["is_active"])

		if err != nil {
			log.Printf("Warning: Failed to seed availability: %v", err)
		}
	}

	log.Printf("   Created %d availability slots", len(availabilities))
}

func seedBookings(db *sql.DB) {
	log.Println("📅 Seeding bookings...")

	// Create a few sample bookings in different states
	tomorrow := time.Now().AddDate(0, 0, 1)
	nextWeek := time.Now().AddDate(0, 0, 7)

	bookings := []map[string]interface{}{
		{
			"id":              "b1111111-1111-1111-1111-111111111111",
			"client_id":       "11111111-1111-1111-1111-111111111111",
			"cleaner_id":      "33333333-3333-3333-3333-333333333333",
			"address_id":      "a1111111-1111-1111-1111-111111111111",
			"status":          "CONFIRMED",
			"service_type":    "STANDARD_CLEANING",
			"scheduled_date":  tomorrow.Format("2006-01-02"),
			"scheduled_time":  "10:00",
			"duration_hours":  3,
			"hourly_rate":     40.00,
			"total_price":     120.00,
			"platform_fee":    12.00,
			"cleaner_payout":  108.00,
		},
		{
			"id":              "b2222222-2222-2222-2222-222222222222",
			"client_id":       "22222222-2222-2222-2222-222222222222",
			"status":          "PENDING",
			"service_type":    "DEEP_CLEANING",
			"scheduled_date":  nextWeek.Format("2006-01-02"),
			"scheduled_time":  "14:00",
			"duration_hours":  5,
			"hourly_rate":     50.00,
			"total_price":     250.00,
			"platform_fee":    25.00,
			"cleaner_payout":  225.00,
		},
	}

	for _, booking := range bookings {
		query := `
			INSERT INTO bookings (
				id, client_id, cleaner_id, address_id, status, service_type,
				scheduled_date, scheduled_time, duration_hours, hourly_rate,
				total_price, platform_fee, cleaner_payout, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`

		_, err := db.Exec(query,
			booking["id"], booking["client_id"], booking["cleaner_id"],
			booking["address_id"], booking["status"], booking["service_type"],
			booking["scheduled_date"], booking["scheduled_time"],
			booking["duration_hours"], booking["hourly_rate"],
			booking["total_price"], booking["platform_fee"], booking["cleaner_payout"])

		if err != nil {
			log.Printf("Warning: Failed to seed booking: %v", err)
		}
	}

	log.Printf("   Created %d bookings", len(bookings))
}

func seedPayments(db *sql.DB) {
	log.Println("💳 Seeding payments...")
	log.Println("   Skipped (requires booking flow)")
}

func seedReviews(db *sql.DB) {
	log.Println("⭐ Seeding reviews...")
	log.Println("   Skipped (requires completed bookings)")
}

func seedDisputes(db *sql.DB) {
	log.Println("⚠️  Seeding disputes...")
	log.Println("   Skipped (requires completed bookings)")
}

func seedCompanies(db *sql.DB) {
	log.Println("🏢 Seeding companies...")
	log.Println("   Skipped (future feature)")
}

func seedPlatformSettings(db *sql.DB) {
	log.Println("⚙️  Seeding platform settings...")

	// Check if settings already exist
	var count int
	db.QueryRow("SELECT COUNT(*) FROM platform_settings").Scan(&count)

	if count > 0 {
		log.Println("   Platform settings already exist, skipping")
		return
	}

	query := `
		INSERT INTO platform_settings (
			platform_fee_percentage, client_cancellation_fee_percentage,
			cleaner_cancellation_fee_percentage, booking_expiration_hours,
			dispute_window_days, created_at, updated_at
		) VALUES (10.0, 20.0, 10.0, 24, 7, NOW(), NOW())
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Warning: Failed to seed platform settings: %v", err)
	} else {
		log.Println("   Created platform settings")
	}
}
