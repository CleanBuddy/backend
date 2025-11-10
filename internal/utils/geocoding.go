package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

// GeocodingService handles address geocoding using Nominatim (OpenStreetMap)
type GeocodingService struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// GeocodingResult represents the response from geocoding API
type GeocodingResult struct {
	Latitude  float64
	Longitude float64
	PlaceID   string
	Address   string
}

// NominatimResponse represents the Nominatim API response
type NominatimResponse struct {
	PlaceID     int64   `json:"place_id"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	DisplayName string  `json:"display_name"`
	Address     Address `json:"address"`
}

// Address represents the structured address from Nominatim
type Address struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	Suburb      string `json:"suburb"`
	City        string `json:"city"`
	County      string `json:"county"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

// NewGeocodingService creates a new geocoding service
// Uses Nominatim (OpenStreetMap) as the free geocoding provider
func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		baseURL: "https://nominatim.openstreetmap.org",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "CleanBuddy/1.0 (contact@cleanbuddy.ro)", // Required by Nominatim usage policy
	}
}

// GeocodeAddress converts an address string to latitude/longitude
// Returns lat, lng, error
func (g *GeocodingService) GeocodeAddress(streetAddress, city, county, country string) (*GeocodingResult, error) {
	// Build the full address query
	addressQuery := fmt.Sprintf("%s, %s, %s, %s", streetAddress, city, county, country)

	// Build the API request URL
	params := url.Values{}
	params.Add("q", addressQuery)
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("limit", "1")
	params.Add("countrycodes", "ro") // Limit to Romania for better results

	requestURL := fmt.Sprintf("%s/search?%s", g.baseURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required User-Agent header (Nominatim usage policy)
	req.Header.Set("User-Agent", g.userAgent)

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	// Parse response
	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check if we got results
	if len(results) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address: %s", addressQuery)
	}

	// Parse the first result
	result := results[0]
	var lat, lng float64

	_, err = fmt.Sscanf(result.Lat, "%f", &lat)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	_, err = fmt.Sscanf(result.Lon, "%f", &lng)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	return &GeocodingResult{
		Latitude:  lat,
		Longitude: lng,
		PlaceID:   fmt.Sprintf("%d", result.PlaceID),
		Address:   result.DisplayName,
	}, nil
}

// CalculateDistance calculates the distance between two points using Haversine formula
// Returns distance in kilometers
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	deltaLatRad := (lat2 - lat1) * math.Pi / 180.0
	deltaLonRad := (lon2 - lon1) * math.Pi / 180.0

	// Haversine formula
	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
