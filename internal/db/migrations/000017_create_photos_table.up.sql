CREATE TABLE IF NOT EXISTS photos (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    booking_id TEXT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    dispute_id TEXT REFERENCES disputes(id) ON DELETE CASCADE,
    uploaded_by TEXT NOT NULL REFERENCES users(id),
    photo_type TEXT NOT NULL CHECK (photo_type IN ('BEFORE', 'AFTER', 'DISPUTE_EVIDENCE')),
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_photos_booking_id ON photos(booking_id);
CREATE INDEX idx_photos_dispute_id ON photos(dispute_id);
CREATE INDEX idx_photos_uploaded_by ON photos(uploaded_by);
CREATE INDEX idx_photos_photo_type ON photos(photo_type);

-- Auto-update updated_at trigger
CREATE TRIGGER update_photos_updated_at
    BEFORE UPDATE ON photos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Constraint: BEFORE/AFTER photos must have booking_id
-- Constraint: DISPUTE_EVIDENCE photos must have dispute_id
ALTER TABLE photos ADD CONSTRAINT check_photo_associations
CHECK (
    (photo_type IN ('BEFORE', 'AFTER') AND booking_id IS NOT NULL) OR
    (photo_type = 'DISPUTE_EVIDENCE' AND dispute_id IS NOT NULL)
);
