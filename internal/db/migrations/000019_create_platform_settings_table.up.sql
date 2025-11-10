-- Platform Settings Table (Single Row Configuration)
CREATE TABLE IF NOT EXISTS platform_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_price NUMERIC(10, 2) NOT NULL DEFAULT 50.00,
    weekend_multiplier NUMERIC(3, 2) NOT NULL DEFAULT 1.20,
    evening_multiplier NUMERIC(3, 2) NOT NULL DEFAULT 1.15,
    platform_fee_percent NUMERIC(5, 2) NOT NULL DEFAULT 10.00,
    email_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    auto_approval_enabled BOOLEAN NOT NULL DEFAULT false,
    maintenance_mode BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create trigger to auto-update updated_at
CREATE TRIGGER update_platform_settings_updated_at
    BEFORE UPDATE ON platform_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default settings (singleton row)
INSERT INTO platform_settings (id, base_price, weekend_multiplier, evening_multiplier, platform_fee_percent)
VALUES ('00000000-0000-0000-0000-000000000001', 50.00, 1.20, 1.15, 10.00)
ON CONFLICT (id) DO NOTHING;
