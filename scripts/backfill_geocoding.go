//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Backfill script to geocode existing addresses and cleaners
// This adds latitude/longitude to records that don't have coordinates yet
// Run with: go run scripts/backfill_geocoding.go

// GeocodingService handles address geocoding using Nominatim
type GeocodingService struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

type GeocodingResult struct {
	Latitude  float64
	Longitude float64
	PlaceID   string
	Address   string
}

type NominatimResponse struct {
	PlaceID     int64  `json:"place_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		baseURL: "https://nominatim.openstreetmap.org",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "CleanBuddy/1.0 (contact@cleanbuddy.ro)",
	}
}

func (g *GeocodingService) GeocodeAddress(streetAddress, city, county, country string) (*GeocodingResult, error) {
	addressQuery := fmt.Sprintf("%s, %s, %s, %s", streetAddress, city, county, country)

	params := url.Values{}
	params.Add("q", addressQuery)
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("limit", "1")
	params.Add("countrycodes", "ro")

	requestURL := fmt.Sprintf("%s/search?%s", g.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", g.userAgent)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nominatim API returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", addressQuery)
	}

	result := results[0]

	var lat, lon float64
	fmt.Sscanf(result.Lat, "%f", &lat)
	fmt.Sscanf(result.Lon, "%f", &lon)

	return &GeocodingResult{
		Latitude:  lat,
		Longitude: lon,
		PlaceID:   fmt.Sprintf("%d", result.PlaceID),
		Address:   result.DisplayName,
	}, nil
}

func main() {
	log.Println("🗺️  Starting geocoding backfill script...")

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
	log.Println("✅ Database connected")

	// Initialize geocoding service
	geocodingService := NewGeocodingService()
	log.Println("✅ Geocoding service initialized (Nominatim)")

	// Backfill addresses
	addressesUpdated, err := backfillAddresses(db, geocodingService)
	if err != nil {
		log.Fatalf("Failed to backfill addresses: %v", err)
	}
	log.Printf("✅ Updated %d addresses with geocoding", addressesUpdated)

	// Backfill cleaners
	cleanersUpdated, err := backfillCleaners(db, geocodingService)
	if err != nil {
		log.Fatalf("Failed to backfill cleaners: %v", err)
	}
	log.Printf("✅ Updated %d cleaners with geocoding", cleanersUpdated)

	log.Println("🎉 Geocoding backfill complete!")
}

// backfillAddresses geocodes all addresses without lat/lng
func backfillAddresses(db *sql.DB, geocodingService *GeocodingService) (int, error) {
	log.Println("📍 Backfilling addresses...")

	// Query addresses without coordinates
	rows, err := db.Query(`
		SELECT id, street_address, city, county, country
		FROM addresses
		WHERE latitude IS NULL OR longitude IS NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query addresses: %w", err)
	}
	defer rows.Close()

	updated := 0
	failed := 0
	total := 0

	for rows.Next() {
		total++
		var id, streetAddress, city, county, country string

		if err := rows.Scan(&id, &streetAddress, &city, &county, &country); err != nil {
			log.Printf("⚠️  Failed to scan address: %v", err)
			failed++
			continue
		}

		// Geocode the address
		result, err := geocodingService.GeocodeAddress(streetAddress, city, county, country)
		if err != nil {
			log.Printf("⚠️  Failed to geocode address %s (%s, %s): %v", id, city, county, err)
			failed++
			// Respect Nominatim rate limit (1 request/second)
			time.Sleep(1 * time.Second)
			continue
		}

		// Update the address with coordinates
		_, err = db.Exec(`
			UPDATE addresses
			SET latitude = $1, longitude = $2, updated_at = NOW()
			WHERE id = $3
		`, result.Latitude, result.Longitude, id)
		if err != nil {
			log.Printf("⚠️  Failed to update address %s: %v", id, err)
			failed++
			continue
		}

		updated++
		log.Printf("  ✅ Address %s: %s, %s → %.6f, %.6f", id, city, county, result.Latitude, result.Longitude)

		// Respect Nominatim rate limit (1 request/second)
		time.Sleep(1 * time.Second)
	}

	if err := rows.Err(); err != nil {
		return updated, fmt.Errorf("error iterating addresses: %w", err)
	}

	log.Printf("📊 Addresses: %d total, %d updated, %d failed", total, updated, failed)
	return updated, nil
}

// backfillCleaners geocodes all cleaners without lat/lng
func backfillCleaners(db *sql.DB, geocodingService *GeocodingService) (int, error) {
	log.Println("👨‍🔧 Backfilling cleaners...")

	// Query cleaners without coordinates
	rows, err := db.Query(`
		SELECT id, street_address, city, county
		FROM cleaners
		WHERE (latitude IS NULL OR longitude IS NULL)
		  AND street_address IS NOT NULL
		  AND city IS NOT NULL
		  AND county IS NOT NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query cleaners: %w", err)
	}
	defer rows.Close()

	updated := 0
	failed := 0
	total := 0

	for rows.Next() {
		total++
		var id string
		var streetAddress, city, county sql.NullString

		if err := rows.Scan(&id, &streetAddress, &city, &county); err != nil {
			log.Printf("⚠️  Failed to scan cleaner: %v", err)
			failed++
			continue
		}

		// Skip if any field is null
		if !streetAddress.Valid || !city.Valid || !county.Valid {
			failed++
			continue
		}

		// Geocode the cleaner's address
		result, err := geocodingService.GeocodeAddress(
			streetAddress.String,
			city.String,
			county.String,
			"România",
		)
		if err != nil {
			log.Printf("⚠️  Failed to geocode cleaner %s (%s, %s): %v", id, city.String, county.String, err)
			failed++
			// Respect Nominatim rate limit (1 request/second)
			time.Sleep(1 * time.Second)
			continue
		}

		// Update the cleaner with coordinates
		_, err = db.Exec(`
			UPDATE cleaners
			SET latitude = $1, longitude = $2, updated_at = NOW()
			WHERE id = $3
		`, result.Latitude, result.Longitude, id)
		if err != nil {
			log.Printf("⚠️  Failed to update cleaner %s: %v", id, err)
			failed++
			continue
		}

		updated++
		log.Printf("  ✅ Cleaner %s: %s, %s → %.6f, %.6f", id, city.String, county.String, result.Latitude, result.Longitude)

		// Respect Nominatim rate limit (1 request/second)
		time.Sleep(1 * time.Second)
	}

	if err := rows.Err(); err != nil {
		return updated, fmt.Errorf("error iterating cleaners: %w", err)
	}

	log.Printf("📊 Cleaners: %d total, %d updated, %d failed", total, updated, failed)
	return updated, nil
}
