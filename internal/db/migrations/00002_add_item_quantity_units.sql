-- +goose Up
-- Add quantity and unit_type columns to items table
ALTER TABLE items ADD COLUMN quantity REAL NOT NULL DEFAULT 1.0;
ALTER TABLE items ADD COLUMN unit_type TEXT NOT NULL DEFAULT 'Count' CHECK(unit_type IN ('Count', 'Grams', 'Liters'));

-- Ensure existing items have explicit values
UPDATE items SET quantity = 1.0, unit_type = 'Count' WHERE quantity IS NULL OR unit_type IS NULL;

-- +goose Down
-- Remove quantity and unit_type columns
ALTER TABLE items DROP COLUMN quantity;
ALTER TABLE items DROP COLUMN unit_type;
