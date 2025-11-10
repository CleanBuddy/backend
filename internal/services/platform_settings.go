package services

import (
	"context"
	"errors"

	"github.com/cleanbuddy/backend/internal/models"
)

type PlatformSettingsService struct {
	repo *models.PlatformSettingsRepository
}

func NewPlatformSettingsService(repo *models.PlatformSettingsRepository) *PlatformSettingsService {
	return &PlatformSettingsService{repo: repo}
}

// GetSettings retrieves the platform settings
func (s *PlatformSettingsService) GetSettings(ctx context.Context) (*models.PlatformSettings, error) {
	return s.repo.Get(ctx)
}

// UpdateSettings updates the platform settings (admin only)
func (s *PlatformSettingsService) UpdateSettings(ctx context.Context, input *models.PlatformSettings) (*models.PlatformSettings, error) {
	// Validate input
	if input.BasePrice < 0 {
		return nil, errors.New("base price cannot be negative")
	}
	if input.WeekendMultiplier < 1.0 {
		return nil, errors.New("weekend multiplier must be >= 1.0")
	}
	if input.EveningMultiplier < 1.0 {
		return nil, errors.New("evening multiplier must be >= 1.0")
	}
	if input.PlatformFeePercent < 0 || input.PlatformFeePercent > 100 {
		return nil, errors.New("platform fee must be between 0 and 100")
	}

	return s.repo.Update(ctx, input)
}
