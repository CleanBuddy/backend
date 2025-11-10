package utils

import (
	"time"
)

// RomanianHoliday represents a Romanian public holiday
type RomanianHoliday struct {
	Name        string
	Month       time.Month
	Day         int
	IsMovable   bool // For Easter-based holidays
	Description string
}

// Romanian public holidays (sărbători legale)
var RomanianHolidays = []RomanianHoliday{
	{
		Name:        "Anul Nou",
		Month:       time.January,
		Day:         1,
		IsMovable:   false,
		Description: "New Year's Day",
	},
	{
		Name:        "Anul Nou (ziua a doua)",
		Month:       time.January,
		Day:         2,
		IsMovable:   false,
		Description: "New Year's Day (second day)",
	},
	{
		Name:        "Ziua Unirii Principatelor Române",
		Month:       time.January,
		Day:         24,
		IsMovable:   false,
		Description: "Unification Day",
	},
	{
		Name:        "Vinerea Mare",
		Month:       0, // Calculated based on Easter
		Day:         -2, // 2 days before Easter
		IsMovable:   true,
		Description: "Good Friday",
	},
	{
		Name:        "Paștele",
		Month:       0, // Calculated
		Day:         0, // Easter Sunday
		IsMovable:   true,
		Description: "Easter Sunday (Orthodox)",
	},
	{
		Name:        "Paștele (a doua zi)",
		Month:       0, // Calculated
		Day:         1, // 1 day after Easter
		IsMovable:   true,
		Description: "Easter Monday (Orthodox)",
	},
	{
		Name:        "Ziua Muncii",
		Month:       time.May,
		Day:         1,
		IsMovable:   false,
		Description: "Labour Day",
	},
	{
		Name:        "Ziua Copilului",
		Month:       time.June,
		Day:         1,
		IsMovable:   false,
		Description: "Children's Day",
	},
	{
		Name:        "Rusaliile",
		Month:       0, // Calculated
		Day:         49, // 49 days after Easter (Pentecost)
		IsMovable:   true,
		Description: "Pentecost (Orthodox)",
	},
	{
		Name:        "Rusaliile (a doua zi)",
		Month:       0, // Calculated
		Day:         50, // 50 days after Easter
		IsMovable:   true,
		Description: "Pentecost Monday (Orthodox)",
	},
	{
		Name:        "Adormirea Maicii Domnului",
		Month:       time.August,
		Day:         15,
		IsMovable:   false,
		Description: "Assumption of Mary",
	},
	{
		Name:        "Sfântul Andrei",
		Month:       time.November,
		Day:         30,
		IsMovable:   false,
		Description: "Saint Andrew's Day",
	},
	{
		Name:        "Ziua Națională a României",
		Month:       time.December,
		Day:         1,
		IsMovable:   false,
		Description: "National Day of Romania",
	},
	{
		Name:        "Crăciunul",
		Month:       time.December,
		Day:         25,
		IsMovable:   false,
		Description: "Christmas Day",
	},
	{
		Name:        "Crăciunul (a doua zi)",
		Month:       time.December,
		Day:         26,
		IsMovable:   false,
		Description: "Christmas (second day)",
	},
}

// IsRomanianHoliday checks if a given date is a Romanian public holiday
func IsRomanianHoliday(date time.Time) bool {
	easterDate := CalculateOrthodoxEaster(date.Year())

	for _, holiday := range RomanianHolidays {
		if holiday.IsMovable {
			// Calculate movable holiday date based on Easter
			holidayDate := easterDate.AddDate(0, 0, holiday.Day)
			if isSameDay(date, holidayDate) {
				return true
			}
		} else {
			// Fixed date holiday
			if date.Month() == holiday.Month && date.Day() == holiday.Day {
				return true
			}
		}
	}

	return false
}

// GetRomanianHolidayName returns the name of the holiday if the date is a holiday
func GetRomanianHolidayName(date time.Time) string {
	easterDate := CalculateOrthodoxEaster(date.Year())

	for _, holiday := range RomanianHolidays {
		if holiday.IsMovable {
			holidayDate := easterDate.AddDate(0, 0, holiday.Day)
			if isSameDay(date, holidayDate) {
				return holiday.Name
			}
		} else {
			if date.Month() == holiday.Month && date.Day() == holiday.Day {
				return holiday.Name
			}
		}
	}

	return ""
}

// CalculateOrthodoxEaster calculates Orthodox Easter date for a given year
// Uses the Meeus/Jones/Butcher algorithm for Orthodox Easter
func CalculateOrthodoxEaster(year int) time.Time {
	// Orthodox Easter calculation (Julian calendar based)
	a := year % 4
	b := year % 7
	c := year % 19
	d := (19*c + 15) % 30
	e := (2*a + 4*b - d + 34) % 7
	month := (d + e + 114) / 31
	day := ((d + e + 114) % 31) + 1

	// Convert from Julian to Gregorian calendar
	// Orthodox Easter is 13 days after the calculated date (for 1900-2099)
	easterDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	easterDate = easterDate.AddDate(0, 0, 13)

	return easterDate
}

// GetRomanianHolidaysForYear returns all Romanian holidays for a specific year
func GetRomanianHolidaysForYear(year int) []time.Time {
	easterDate := CalculateOrthodoxEaster(year)
	holidays := []time.Time{}

	for _, holiday := range RomanianHolidays {
		if holiday.IsMovable {
			holidayDate := easterDate.AddDate(0, 0, holiday.Day)
			holidays = append(holidays, holidayDate)
		} else {
			holidayDate := time.Date(year, holiday.Month, holiday.Day, 0, 0, 0, 0, time.UTC)
			holidays = append(holidays, holidayDate)
		}
	}

	return holidays
}

// IsWeekend checks if a date falls on weekend (Saturday or Sunday)
func IsWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// isSameDay checks if two dates are the same day (ignoring time)
func isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// GetNextHoliday returns the next upcoming Romanian holiday from the given date
func GetNextHoliday(fromDate time.Time) (time.Time, string) {
	currentYear := fromDate.Year()

	// Check current year holidays
	for _, holiday := range GetRomanianHolidaysForYear(currentYear) {
		if holiday.After(fromDate) {
			return holiday, GetRomanianHolidayName(holiday)
		}
	}

	// If no holidays left this year, get first holiday of next year
	nextYearHolidays := GetRomanianHolidaysForYear(currentYear + 1)
	if len(nextYearHolidays) > 0 {
		return nextYearHolidays[0], GetRomanianHolidayName(nextYearHolidays[0])
	}

	return time.Time{}, ""
}
