CREATE TABLE checkins (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    booking_id TEXT NOT NULL REFERENCES bookings(id),
    cleaner_id TEXT NOT NULL REFERENCES cleaners(id),
    check_in_time TIMESTAMP WITH TIME ZONE,
    check_in_latitude DECIMAL(10, 8),
    check_in_longitude DECIMAL(11, 8),
    check_out_time TIMESTAMP WITH TIME ZONE,
    check_out_latitude DECIMAL(10, 8),
    check_out_longitude DECIMAL(11, 8),
    total_hours_worked DECIMAL(4, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_checkins_booking_id ON checkins(booking_id);
CREATE INDEX idx_checkins_cleaner_id ON checkins(cleaner_id);
