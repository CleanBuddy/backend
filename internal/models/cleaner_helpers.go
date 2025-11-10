package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ParseSpecializations parses the JSONB specializations field into a string slice
func (c *Cleaner) ParseSpecializations() ([]string, error) {
	if c.Specializations == nil || len(c.Specializations) == 0 {
		return []string{}, nil
	}

	var specializations []string
	if err := json.Unmarshal(c.Specializations, &specializations); err != nil {
		return nil, fmt.Errorf("failed to parse specializations: %w", err)
	}

	return specializations, nil
}

// ParseLanguages parses the JSONB languages field into a string slice
func (c *Cleaner) ParseLanguages() ([]string, error) {
	if c.Languages == nil || len(c.Languages) == 0 {
		return []string{}, nil
	}

	var languages []string
	if err := json.Unmarshal(c.Languages, &languages); err != nil {
		return nil, fmt.Errorf("failed to parse languages: %w", err)
	}

	return languages, nil
}

// SpecializationsArray is a custom type for JSONB array
type SpecializationsArray []string

// Value implements the driver.Valuer interface for database storage
func (s SpecializationsArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *SpecializationsArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SpecializationsArray: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, s)
}

// LanguagesArray is a custom type for JSONB array
type LanguagesArray []string

// Value implements the driver.Valuer interface for database storage
func (l LanguagesArray) Value() (driver.Value, error) {
	if len(l) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(l)
}

// Scan implements the sql.Scanner interface for database retrieval
func (l *LanguagesArray) Scan(value interface{}) error {
	if value == nil {
		*l = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan LanguagesArray: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, l)
}
