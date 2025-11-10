package utils

import (
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
		delta    float64 // Acceptable error margin in km
	}{
		{
			name:     "Bucharest to Cluj-Napoca",
			lat1:     44.4268, // Bucharest
			lon1:     26.1025,
			lat2:     46.7712, // Cluj-Napoca
			lon2:     23.6236,
			expected: 320.0, // Approximately 320 km
			delta:    10.0,
		},
		{
			name:     "Same location (zero distance)",
			lat1:     44.4268,
			lon1:     26.1025,
			lat2:     44.4268,
			lon2:     26.1025,
			expected: 0.0,
			delta:    0.1,
		},
		{
			name:     "Bucharest to Timișoara",
			lat1:     44.4268, // Bucharest
			lon1:     26.1025,
			lat2:     45.7489, // Timișoara
			lon2:     21.2087,
			expected: 420.0, // Approximately 420 km
			delta:    10.0,
		},
		{
			name:     "Short distance within Bucharest",
			lat1:     44.4268, // Piața Unirii
			lon1:     26.1025,
			lat2:     44.4355, // Piața Victoriei (approx 1 km)
			lon2:     26.0997,
			expected: 1.0,
			delta:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			// Check if distance is within acceptable range
			diff := distance - tt.expected
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.delta {
				t.Errorf("CalculateDistance() = %.2f km, expected approximately %.2f km (±%.2f km), difference: %.2f km",
					distance, tt.expected, tt.delta, diff)
			}
		})
	}
}

func TestGeocodeAddress(t *testing.T) {
	// Skip if not in integration test mode (to avoid hitting Nominatim API in regular tests)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := NewGeocodingService()

	tests := []struct {
		name          string
		streetAddress string
		city          string
		county        string
		country       string
		expectedLat   float64 // Approximate expected latitude
		expectedLon   float64 // Approximate expected longitude
		delta         float64 // Acceptable error margin in degrees
		shouldError   bool
	}{
		{
			name:          "Valid Bucharest address",
			streetAddress: "Bulevardul Unirii 1",
			city:          "București",
			county:        "București",
			country:       "România",
			expectedLat:   44.427,
			expectedLon:   26.103,
			delta:         0.01, // About 1 km tolerance
			shouldError:   false,
		},
		{
			name:          "Valid Cluj-Napoca address",
			streetAddress: "Piața Unirii 1",
			city:          "Cluj-Napoca",
			county:        "Cluj",
			country:       "România",
			expectedLat:   46.771,
			expectedLon:   23.590,
			delta:         0.01,
			shouldError:   false,
		},
		{
			name:          "Invalid address",
			streetAddress: "NonexistentStreet 999999",
			city:          "NonexistentCity",
			county:        "NonexistentCounty",
			country:       "România",
			expectedLat:   0.0,
			expectedLon:   0.0,
			delta:         0.0,
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GeocodeAddress(tt.streetAddress, tt.city, tt.county, tt.country)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Check latitude
			latDiff := result.Latitude - tt.expectedLat
			if latDiff < 0 {
				latDiff = -latDiff
			}
			if latDiff > tt.delta {
				t.Errorf("Latitude = %.6f, expected approximately %.6f (±%.6f), difference: %.6f",
					result.Latitude, tt.expectedLat, tt.delta, latDiff)
			}

			// Check longitude
			lonDiff := result.Longitude - tt.expectedLon
			if lonDiff < 0 {
				lonDiff = -lonDiff
			}
			if lonDiff > tt.delta {
				t.Errorf("Longitude = %.6f, expected approximately %.6f (±%.6f), difference: %.6f",
					result.Longitude, tt.expectedLon, tt.delta, lonDiff)
			}

			// Check that PlaceID and Address are populated
			if result.PlaceID == "" {
				t.Error("PlaceID should not be empty")
			}
			if result.Address == "" {
				t.Error("Address should not be empty")
			}

			t.Logf("Geocoded address: %s", result.Address)
			t.Logf("Coordinates: %.6f, %.6f", result.Latitude, result.Longitude)
		})
	}
}
