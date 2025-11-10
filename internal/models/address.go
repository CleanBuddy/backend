package models

import (
	"database/sql"
	"time"
)

// Address represents a user's address in the system
type Address struct {
	ID             string
	UserID         string
	Label          string
	StreetAddress  string
	Apartment      sql.NullString
	City           string
	County         string
	PostalCode     sql.NullString
	Country        string
	AdditionalInfo sql.NullString
	IsDefault      bool
	Latitude       sql.NullFloat64 // Decimal degrees (-90 to 90)
	Longitude      sql.NullFloat64 // Decimal degrees (-180 to 180)
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AddressRepository handles address database operations
type AddressRepository struct {
	db *sql.DB
}

// NewAddressRepository creates a new address repository
func NewAddressRepository(db *sql.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// GetByUserID gets all addresses for a user
func (r *AddressRepository) GetByUserID(userID string) ([]*Address, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, label, street_address, apartment, city, county,
		       postal_code, country, additional_info, is_default, latitude, longitude,
		       created_at, updated_at
		FROM addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*Address
	for rows.Next() {
		address := &Address{}
		err := rows.Scan(
			&address.ID, &address.UserID, &address.Label, &address.StreetAddress,
			&address.Apartment, &address.City, &address.County, &address.PostalCode,
			&address.Country, &address.AdditionalInfo, &address.IsDefault,
			&address.Latitude, &address.Longitude,
			&address.CreatedAt, &address.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetByID gets an address by ID
func (r *AddressRepository) GetByID(id string) (*Address, error) {
	address := &Address{}
	err := r.db.QueryRow(`
		SELECT id, user_id, label, street_address, apartment, city, county,
		       postal_code, country, additional_info, is_default, latitude, longitude,
		       created_at, updated_at
		FROM addresses
		WHERE id = $1
	`, id).Scan(
		&address.ID, &address.UserID, &address.Label, &address.StreetAddress,
		&address.Apartment, &address.City, &address.County, &address.PostalCode,
		&address.Country, &address.AdditionalInfo, &address.IsDefault,
		&address.Latitude, &address.Longitude,
		&address.CreatedAt, &address.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return address, nil
}

// Create creates a new address
func (r *AddressRepository) Create(address *Address) error {
	return r.db.QueryRow(`
		INSERT INTO addresses (user_id, label, street_address, apartment, city, county, postal_code, country, additional_info, is_default, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`, address.UserID, address.Label, address.StreetAddress, address.Apartment,
		address.City, address.County, address.PostalCode, address.Country,
		address.AdditionalInfo, address.IsDefault, address.Latitude, address.Longitude).
		Scan(&address.ID, &address.CreatedAt, &address.UpdatedAt)
}

// Update updates an address
func (r *AddressRepository) Update(address *Address) error {
	_, err := r.db.Exec(`
		UPDATE addresses
		SET label = $2, street_address = $3, apartment = $4, city = $5,
		    county = $6, postal_code = $7, country = $8, additional_info = $9, is_default = $10,
		    latitude = $11, longitude = $12, updated_at = NOW()
		WHERE id = $1
	`, address.ID, address.Label, address.StreetAddress, address.Apartment,
		address.City, address.County, address.PostalCode, address.Country,
		address.AdditionalInfo, address.IsDefault, address.Latitude, address.Longitude)
	return err
}

// Delete deletes an address
func (r *AddressRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM addresses WHERE id = $1`, id)
	return err
}
