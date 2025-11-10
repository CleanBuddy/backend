package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// GeocodingService handles address geocoding
type GeocodingService struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeocodingService creates a new geocoding service
func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		apiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GeocodeResult represents geocoding API response
type GeocodeResult struct {
	Latitude  float64
	Longitude float64
	FormattedAddress string
}

// GoogleMapsGeocodeResponse represents the Google Maps API response
type GoogleMapsGeocodeResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

// Geocode converts an address string to latitude/longitude coordinates
func (s *GeocodingService) Geocode(street, city, county, postalCode, country string) (*GeocodeResult, error) {
	// If API key is not set, return mock coordinates for development
	if s.apiKey == "" || os.Getenv("ENV") == "development" {
		return s.mockGeocode(city), nil
	}

	// Build full address string
	address := fmt.Sprintf("%s, %s, %s %s, %s", street, city, county, postalCode, country)

	// Call Google Maps Geocoding API
	return s.geocodeWithGoogle(address)
}

// geocodeWithGoogle calls the Google Maps Geocoding API
func (s *GeocodingService) geocodeWithGoogle(address string) (*GeocodeResult, error) {
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"

	// Build request URL
	params := url.Values{}
	params.Add("address", address)
	params.Add("key", s.apiKey)
	params.Add("region", "ro") // Bias results to Romania

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make HTTP request
	resp, err := s.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("geocoding API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read geocoding response: %w", err)
	}

	// Parse JSON response
	var result GoogleMapsGeocodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check status
	if result.Status != "OK" {
		return nil, fmt.Errorf("geocoding failed with status: %s", result.Status)
	}

	// Extract first result
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address")
	}

	firstResult := result.Results[0]
	return &GeocodeResult{
		Latitude:         firstResult.Geometry.Location.Lat,
		Longitude:        firstResult.Geometry.Location.Lng,
		FormattedAddress: firstResult.FormattedAddress,
	}, nil
}

// mockGeocode returns mock coordinates for development/testing
// These are approximate city center coordinates for Romanian cities
func (s *GeocodingService) mockGeocode(city string) *GeocodeResult {
	// Mock coordinates for major Romanian cities
	mockCoordinates := map[string]GeocodeResult{
		"București": {Latitude: 44.4268, Longitude: 26.1025, FormattedAddress: "București, Romania"},
		"Bucharest": {Latitude: 44.4268, Longitude: 26.1025, FormattedAddress: "București, Romania"},
		"Cluj-Napoca": {Latitude: 46.7712, Longitude: 23.6236, FormattedAddress: "Cluj-Napoca, Romania"},
		"Timișoara": {Latitude: 45.7489, Longitude: 21.2087, FormattedAddress: "Timișoara, Romania"},
		"Iași": {Latitude: 47.1585, Longitude: 27.6014, FormattedAddress: "Iași, Romania"},
		"Constanța": {Latitude: 44.1598, Longitude: 28.6348, FormattedAddress: "Constanța, Romania"},
		"Craiova": {Latitude: 44.3302, Longitude: 23.7949, FormattedAddress: "Craiova, Romania"},
		"Brașov": {Latitude: 45.6527, Longitude: 25.6102, FormattedAddress: "Brașov, Romania"},
		"Galați": {Latitude: 45.4353, Longitude: 28.0080, FormattedAddress: "Galați, Romania"},
		"Ploiești": {Latitude: 44.9401, Longitude: 26.0266, FormattedAddress: "Ploiești, Romania"},
	}

	// Return mock coordinates if city is found
	if coords, ok := mockCoordinates[city]; ok {
		return &coords
	}

	// Default to Bucharest center if city not found
	defaultCoords := GeocodeResult{
		Latitude:         44.4268,
		Longitude:        26.1025,
		FormattedAddress: "București, Romania (default)",
	}
	return &defaultCoords
}
