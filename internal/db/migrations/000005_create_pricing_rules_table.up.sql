-- Pricing rules table for dynamic pricing
CREATE TABLE IF NOT EXISTS pricing_rules (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,

    -- Rule details
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Service types
    service_type VARCHAR(50) NOT NULL,

    -- Base pricing
    base_price_per_hour DECIMAL(10, 2) NOT NULL,
    minimum_hours INTEGER NOT NULL DEFAULT 2,

    -- Area-based pricing (sqm = square meters)
    price_per_sqm DECIMAL(10, 2),

    -- Add-ons pricing
    deep_cleaning_multiplier DECIMAL(3, 2) DEFAULT 1.5,
    window_cleaning_price DECIMAL(10, 2) DEFAULT 0.00,
    carpet_cleaning_price_per_sqm DECIMAL(10, 2) DEFAULT 0.00,

    -- Time-based pricing
    weekend_multiplier DECIMAL(3, 2) DEFAULT 1.2,
    evening_multiplier DECIMAL(3, 2) DEFAULT 1.1,

    -- Platform fees
    platform_fee_percentage DECIMAL(5, 2) NOT NULL DEFAULT 10.00,
    first_booking_discount_percentage DECIMAL(5, 2) DEFAULT 0.00,

    -- Activity status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes
CREATE INDEX idx_pricing_rules_service_type ON pricing_rules(service_type);
CREATE INDEX idx_pricing_rules_is_active ON pricing_rules(is_active);

-- Add trigger for updated_at
CREATE TRIGGER set_pricing_rules_updated_at
    BEFORE UPDATE ON pricing_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint for service_type enum
ALTER TABLE pricing_rules
    ADD CONSTRAINT pricing_rules_service_type_check
    CHECK (service_type IN ('STANDARD', 'DEEP_CLEANING', 'OFFICE', 'POST_RENOVATION', 'MOVE_IN_OUT'));

-- Insert default pricing rules
INSERT INTO pricing_rules (name, description, service_type, base_price_per_hour, minimum_hours, price_per_sqm, platform_fee_percentage, first_booking_discount_percentage)
VALUES
    ('Curățenie Standard', 'Curățenie generală pentru locuințe', 'STANDARD', 50.00, 2, 2.50, 10.00, 10.00),
    ('Curățenie Profundă', 'Curățenie profundă cu detalii', 'DEEP_CLEANING', 60.00, 3, 3.00, 10.00, 10.00),
    ('Curățenie Birouri', 'Curățenie pentru spații de birouri', 'OFFICE', 45.00, 2, 2.00, 10.00, 0.00),
    ('După Renovare', 'Curățenie după lucrări de renovare', 'POST_RENOVATION', 70.00, 4, 4.00, 10.00, 5.00),
    ('Mutare', 'Curățenie la mutare (intrare/ieșire)', 'MOVE_IN_OUT', 65.00, 3, 3.50, 10.00, 5.00);

-- Comment on table
COMMENT ON TABLE pricing_rules IS 'Dynamic pricing rules for different service types';
