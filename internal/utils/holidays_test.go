package utils

import (
	"testing"
	"time"
)

func TestIsRomanianHoliday(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
		holiday  string
	}{
		{
			name:     "New Year's Day 2025",
			date:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
			holiday:  "Anul Nou",
		},
		{
			name:     "Christmas 2025",
			date:     time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: true,
			holiday:  "Crăciunul",
		},
		{
			name:     "National Day 2025",
			date:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
			holiday:  "Ziua Națională a României",
		},
		{
			name:     "Regular day",
			date:     time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			expected: false,
			holiday:  "",
		},
		{
			name:     "Labour Day",
			date:     time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
			holiday:  "Ziua Muncii",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRomanianHoliday(tt.date)
			if result != tt.expected {
				t.Errorf("IsRomanianHoliday(%v) = %v; want %v", tt.date, result, tt.expected)
			}

			if tt.expected {
				holidayName := GetRomanianHolidayName(tt.date)
				if holidayName != tt.holiday {
					t.Errorf("GetRomanianHolidayName(%v) = %v; want %v", tt.date, holidayName, tt.holiday)
				}
			}
		})
	}
}

func TestCalculateOrthodoxEaster(t *testing.T) {
	// Known Orthodox Easter dates
	tests := []struct {
		year     int
		expected time.Time
	}{
		{2024, time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC)},  // May 5, 2024
		{2025, time.Date(2025, 4, 20, 0, 0, 0, 0, time.UTC)}, // April 20, 2025
		{2026, time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)}, // April 12, 2026
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.year)), func(t *testing.T) {
			result := CalculateOrthodoxEaster(tt.year)
			if !isSameDay(result, tt.expected) {
				t.Errorf("CalculateOrthodoxEaster(%d) = %v; want %v", tt.year, result, tt.expected)
			}
		})
	}
}

func TestGetRomanianHolidaysForYear(t *testing.T) {
	holidays := GetRomanianHolidaysForYear(2025)

	// Should have 15 holidays
	if len(holidays) != 15 {
		t.Errorf("GetRomanianHolidaysForYear(2025) returned %d holidays; want 15", len(holidays))
	}

	// Check that New Year is included
	hasNewYear := false
	for _, h := range holidays {
		if h.Month() == time.January && h.Day() == 1 {
			hasNewYear = true
			break
		}
	}
	if !hasNewYear {
		t.Error("New Year's Day not found in 2025 holidays")
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Saturday",
			date:     time.Date(2025, 10, 11, 0, 0, 0, 0, time.UTC), // Saturday
			expected: true,
		},
		{
			name:     "Sunday",
			date:     time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC), // Sunday
			expected: true,
		},
		{
			name:     "Monday",
			date:     time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC), // Monday
			expected: false,
		},
		{
			name:     "Wednesday",
			date:     time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), // Wednesday
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekend(tt.date)
			if result != tt.expected {
				t.Errorf("IsWeekend(%v) = %v; want %v", tt.date, result, tt.expected)
			}
		})
	}
}
