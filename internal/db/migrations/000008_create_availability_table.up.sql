-- Create availability table for cleaner schedules
CREATE TABLE IF NOT EXISTS availability (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Relations
    cleaner_id TEXT NOT NULL REFERENCES cleaners(id) ON DELETE CASCADE,

    -- Schedule type
    type VARCHAR(20) NOT NULL, -- 'RECURRING', 'ONE_TIME', 'BLOCKED'

    -- Day of week for recurring slots (0 = Sunday, 6 = Saturday)
    day_of_week INTEGER, -- NULL for one-time slots

    -- Specific date for one-time slots
    specific_date DATE, -- NULL for recurring slots

    -- Time slots
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Notes
    notes TEXT,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_availability_cleaner_id ON availability(cleaner_id);
CREATE INDEX idx_availability_day_of_week ON availability(day_of_week);
CREATE INDEX idx_availability_specific_date ON availability(specific_date);
CREATE INDEX idx_availability_is_active ON availability(is_active);

-- Trigger for updated_at
CREATE TRIGGER update_availability_updated_at
    BEFORE UPDATE ON availability
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Constraints
-- Ensure day_of_week is valid (0-6)
ALTER TABLE availability ADD CONSTRAINT check_day_of_week
    CHECK (day_of_week IS NULL OR (day_of_week >= 0 AND day_of_week <= 6));

-- Ensure start_time < end_time
ALTER TABLE availability ADD CONSTRAINT check_time_range
    CHECK (start_time < end_time);

-- Ensure either day_of_week or specific_date is set (not both, not neither)
ALTER TABLE availability ADD CONSTRAINT check_schedule_type
    CHECK (
        (type = 'RECURRING' AND day_of_week IS NOT NULL AND specific_date IS NULL) OR
        (type = 'ONE_TIME' AND specific_date IS NOT NULL AND day_of_week IS NULL) OR
        (type = 'BLOCKED' AND specific_date IS NOT NULL AND day_of_week IS NULL)
    );

-- Comments
COMMENT ON TABLE availability IS 'Cleaner availability schedules (recurring weekly + one-time slots + blocked dates)';
COMMENT ON COLUMN availability.type IS 'Schedule type: RECURRING (weekly), ONE_TIME (specific date), BLOCKED (unavailable)';
COMMENT ON COLUMN availability.day_of_week IS 'Day of week (0=Sunday, 6=Saturday) for RECURRING type';
COMMENT ON COLUMN availability.specific_date IS 'Specific date for ONE_TIME or BLOCKED type';
