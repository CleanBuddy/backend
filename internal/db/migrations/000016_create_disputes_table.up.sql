-- Disputes table for quality complaints and resolutions
CREATE TABLE IF NOT EXISTS disputes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    booking_id TEXT NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
    created_by TEXT NOT NULL REFERENCES users(id),
    assigned_to TEXT REFERENCES users(id), -- Admin handling the dispute

    -- Dispute details
    dispute_type TEXT NOT NULL CHECK (dispute_type IN ('QUALITY_ISSUE', 'DAMAGE', 'NO_SHOW', 'PRICING', 'OTHER')),
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'UNDER_REVIEW', 'RESOLVED', 'CLOSED')),
    description TEXT NOT NULL,

    -- Resolution
    resolution_type TEXT CHECK (resolution_type IN ('PARTIAL_REFUND', 'FULL_REFUND', 'RECLEAN', 'REJECTED')),
    resolution_notes TEXT,
    refund_amount DECIMAL(10, 2),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by TEXT REFERENCES users(id),

    -- Cleaner response
    cleaner_response TEXT,
    cleaner_responded_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_disputes_booking_id ON disputes(booking_id);
CREATE INDEX idx_disputes_created_by ON disputes(created_by);
CREATE INDEX idx_disputes_status ON disputes(status);
CREATE INDEX idx_disputes_assigned_to ON disputes(assigned_to);
CREATE INDEX idx_disputes_created_at ON disputes(created_at DESC);

-- Trigger to update updated_at
CREATE TRIGGER set_disputes_updated_at
    BEFORE UPDATE ON disputes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE disputes IS 'Customer disputes for completed bookings requiring admin resolution';
COMMENT ON COLUMN disputes.dispute_type IS 'Type of dispute: QUALITY_ISSUE, DAMAGE, NO_SHOW, PRICING, OTHER';
COMMENT ON COLUMN disputes.status IS 'Dispute lifecycle: OPEN → UNDER_REVIEW → RESOLVED → CLOSED';
COMMENT ON COLUMN disputes.resolution_type IS 'Admin decision: PARTIAL_REFUND, FULL_REFUND, RECLEAN, REJECTED';
