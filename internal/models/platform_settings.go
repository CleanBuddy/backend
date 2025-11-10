package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PlatformSettings struct {
	ID                        uuid.UUID `db:"id"`
	BasePrice                 float64   `db:"base_price"`
	WeekendMultiplier         float64   `db:"weekend_multiplier"`
	EveningMultiplier         float64   `db:"evening_multiplier"`
	PlatformFeePercent        float64   `db:"platform_fee_percent"`
	EmailNotificationsEnabled bool      `db:"email_notifications_enabled"`
	AutoApprovalEnabled       bool      `db:"auto_approval_enabled"`
	MaintenanceMode           bool      `db:"maintenance_mode"`
	CreatedAt                 time.Time `db:"created_at"`
	UpdatedAt                 time.Time `db:"updated_at"`
}

type PlatformSettingsRepository struct {
	db *sql.DB
}

func NewPlatformSettingsRepository(db *sql.DB) *PlatformSettingsRepository {
	return &PlatformSettingsRepository{db: db}
}

// Get returns the platform settings (singleton row)
func (r *PlatformSettingsRepository) Get(ctx context.Context) (*PlatformSettings, error) {
	var settings PlatformSettings
	query := `
		SELECT id, base_price, weekend_multiplier, evening_multiplier,
		       platform_fee_percent, email_notifications_enabled,
		       auto_approval_enabled, maintenance_mode, created_at, updated_at
		FROM platform_settings
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query).Scan(
		&settings.ID,
		&settings.BasePrice,
		&settings.WeekendMultiplier,
		&settings.EveningMultiplier,
		&settings.PlatformFeePercent,
		&settings.EmailNotificationsEnabled,
		&settings.AutoApprovalEnabled,
		&settings.MaintenanceMode,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Update updates the platform settings
func (r *PlatformSettingsRepository) Update(ctx context.Context, input *PlatformSettings) (*PlatformSettings, error) {
	query := `
		UPDATE platform_settings
		SET base_price = $1,
		    weekend_multiplier = $2,
		    evening_multiplier = $3,
		    platform_fee_percent = $4,
		    email_notifications_enabled = $5,
		    auto_approval_enabled = $6,
		    maintenance_mode = $7,
		    updated_at = NOW()
		WHERE id = $8
		RETURNING id, base_price, weekend_multiplier, evening_multiplier,
		          platform_fee_percent, email_notifications_enabled,
		          auto_approval_enabled, maintenance_mode, created_at, updated_at
	`

	var settings PlatformSettings
	err := r.db.QueryRowContext(ctx, query,
		input.BasePrice,
		input.WeekendMultiplier,
		input.EveningMultiplier,
		input.PlatformFeePercent,
		input.EmailNotificationsEnabled,
		input.AutoApprovalEnabled,
		input.MaintenanceMode,
		input.ID,
	).Scan(
		&settings.ID,
		&settings.BasePrice,
		&settings.WeekendMultiplier,
		&settings.EveningMultiplier,
		&settings.PlatformFeePercent,
		&settings.EmailNotificationsEnabled,
		&settings.AutoApprovalEnabled,
		&settings.MaintenanceMode,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}
