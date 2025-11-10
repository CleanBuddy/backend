-- Add missing fields from booking draft to bookings table

-- Add frequency for recurring bookings
ALTER TABLE bookings
    ADD COLUMN frequency VARCHAR(20);

ALTER TABLE bookings
    ADD CONSTRAINT bookings_frequency_check
    CHECK (frequency IN ('one_time', 'weekly', 'biweekly', 'monthly'));

COMMENT ON COLUMN bookings.frequency IS 'Booking frequency: one_time, weekly, biweekly, monthly';

-- Add time preferences as JSONB (array of {date, timeSlots})
ALTER TABLE bookings
    ADD COLUMN time_preferences JSONB;

COMMENT ON COLUMN bookings.time_preferences IS 'Client available time slots: [{date: "2024-01-15", timeSlots: ["morning", "afternoon"]}]';

-- Add extras as array
ALTER TABLE bookings
    ADD COLUMN extras TEXT[];

COMMENT ON COLUMN bookings.extras IS 'Extra services: {ironing, windows, fridge_oven, balcony}';

-- Add supplies field
ALTER TABLE bookings
    ADD COLUMN supplies VARCHAR(20);

ALTER TABLE bookings
    ADD CONSTRAINT bookings_supplies_check
    CHECK (supplies IN ('client_provides', 'cleaner_provides'));

COMMENT ON COLUMN bookings.supplies IS 'Who provides cleaning supplies';
