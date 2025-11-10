CREATE TABLE IF NOT EXISTS payouts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    cleaner_id TEXT NOT NULL REFERENCES users(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'SENT', 'FAILED', 'CANCELLED')),
    total_bookings INTEGER NOT NULL DEFAULT 0,
    total_earnings DECIMAL(10, 2) NOT NULL DEFAULT 0,
    platform_fees DECIMAL(10, 2) NOT NULL DEFAULT 0,
    net_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    iban TEXT,
    transfer_reference TEXT,
    settlement_invoice_url TEXT,
    paid_at TIMESTAMP WITH TIME ZONE,
    failed_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(cleaner_id, period_start, period_end)
);

-- Indexes for performance
CREATE INDEX idx_payouts_cleaner_id ON payouts(cleaner_id);
CREATE INDEX idx_payouts_status ON payouts(status);
CREATE INDEX idx_payouts_period ON payouts(period_start, period_end);
CREATE INDEX idx_payouts_paid_at ON payouts(paid_at);

-- Auto-update updated_at trigger
CREATE TRIGGER update_payouts_updated_at
    BEFORE UPDATE ON payouts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Payout line items (for detailed breakdown)
CREATE TABLE IF NOT EXISTS payout_line_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    payout_id TEXT NOT NULL REFERENCES payouts(id) ON DELETE CASCADE,
    booking_id TEXT NOT NULL REFERENCES bookings(id),
    booking_date TIMESTAMP WITH TIME ZONE NOT NULL,
    service_type TEXT NOT NULL,
    booking_amount DECIMAL(10, 2) NOT NULL,
    platform_fee_rate DECIMAL(5, 2) NOT NULL,
    platform_fee DECIMAL(10, 2) NOT NULL,
    cleaner_earnings DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_payout_line_items_payout_id ON payout_line_items(payout_id);
CREATE INDEX idx_payout_line_items_booking_id ON payout_line_items(booking_id);
