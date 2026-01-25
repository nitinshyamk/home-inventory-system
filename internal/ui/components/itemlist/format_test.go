package itemlist

import (
	"testing"

	"home-inventory-system/internal/domain"
)

func TestFormatQuantity(t *testing.T) {
	tests := []struct {
		name     string
		quantity float64
		unitType domain.UnitType
		want     string
	}{
		// Count tests - should round to nearest integer
		{
			name:     "count - whole number",
			quantity: 5.0,
			unitType: domain.UnitTypeCount,
			want:     "5",
		},
		{
			name:     "count - rounds down",
			quantity: 5.4,
			unitType: domain.UnitTypeCount,
			want:     "5",
		},
		{
			name:     "count - rounds up",
			quantity: 5.6,
			unitType: domain.UnitTypeCount,
			want:     "6",
		},
		{
			name:     "count - exactly 0.5 rounds up",
			quantity: 5.5,
			unitType: domain.UnitTypeCount,
			want:     "6",
		},
		{
			name:     "count - single item",
			quantity: 1.0,
			unitType: domain.UnitTypeCount,
			want:     "1",
		},
		{
			name:     "count - zero",
			quantity: 0.0,
			unitType: domain.UnitTypeCount,
			want:     "0",
		},

		// Grams tests - converts to oz, then to lbs if >= 16 oz
		{
			name:     "grams - small amount in oz",
			quantity: 28.3495, // Exactly 1 oz
			unitType: domain.UnitTypeGrams,
			want:     "1.0 oz",
		},
		{
			name:     "grams - 50g to oz",
			quantity: 50.0,
			unitType: domain.UnitTypeGrams,
			want:     "1.8 oz",
		},
		{
			name:     "grams - 100g to oz",
			quantity: 100.0,
			unitType: domain.UnitTypeGrams,
			want:     "3.5 oz",
		},
		{
			name:     "grams - exactly 16 oz (1 lb)",
			quantity: 453.592, // 16 oz in grams
			unitType: domain.UnitTypeGrams,
			want:     "1.0 lbs",
		},
		{
			name:     "grams - 500g to lbs",
			quantity: 500.0,
			unitType: domain.UnitTypeGrams,
			want:     "1.1 lbs",
		},
		{
			name:     "grams - 1000g (1kg) to lbs",
			quantity: 1000.0,
			unitType: domain.UnitTypeGrams,
			want:     "2.2 lbs",
		},
		{
			name:     "grams - 2000g to lbs",
			quantity: 2000.0,
			unitType: domain.UnitTypeGrams,
			want:     "4.4 lbs",
		},
		{
			name:     "grams - 5000g to lbs",
			quantity: 5000.0,
			unitType: domain.UnitTypeGrams,
			want:     "11.0 lbs",
		},
		{
			name:     "grams - 15 oz (stays in oz)",
			quantity: 425.0,
			unitType: domain.UnitTypeGrams,
			want:     "15.0 oz",
		},

		// Liters tests - converts to fl oz, then to gal if >= 128 fl oz (1 gallon)
		{
			name:     "liters - small amount in fl oz",
			quantity: 0.5,
			unitType: domain.UnitTypeLiters,
			want:     "16.9 fl oz",
		},
		{
			name:     "liters - 1 liter to fl oz",
			quantity: 1.0,
			unitType: domain.UnitTypeLiters,
			want:     "33.8 fl oz",
		},
		{
			name:     "liters - 2 liters to fl oz",
			quantity: 2.0,
			unitType: domain.UnitTypeLiters,
			want:     "67.6 fl oz",
		},
		{
			name:     "liters - 3 liters to fl oz",
			quantity: 3.0,
			unitType: domain.UnitTypeLiters,
			want:     "101.4 fl oz",
		},
		{
			name:     "liters - exactly 1 gallon",
			quantity: 3.78541, // 1 gallon in liters (128.0 fl oz)
			unitType: domain.UnitTypeLiters,
			want:     "128.0 fl oz", // Displays as fl oz when exactly 128.0
		},
		{
			name:     "liters - 4 liters to gal",
			quantity: 4.0,
			unitType: domain.UnitTypeLiters,
			want:     "1.1 gal",
		},
		{
			name:     "liters - 5 liters to gal",
			quantity: 5.0,
			unitType: domain.UnitTypeLiters,
			want:     "1.3 gal",
		},
		{
			name:     "liters - 10 liters to gal",
			quantity: 10.0,
			unitType: domain.UnitTypeLiters,
			want:     "2.6 gal",
		},
		{
			name:     "liters - 3.75L (stays in fl oz)",
			quantity: 3.75, // 126.8 fl oz
			unitType: domain.UnitTypeLiters,
			want:     "126.8 fl oz",
		},

		// Edge cases
		{
			name:     "count - very large number",
			quantity: 1000.0,
			unitType: domain.UnitTypeCount,
			want:     "1000",
		},
		{
			name:     "grams - very small amount",
			quantity: 1.0,
			unitType: domain.UnitTypeGrams,
			want:     "0.0 oz",
		},
		{
			name:     "liters - very small amount",
			quantity: 0.01,
			unitType: domain.UnitTypeLiters,
			want:     "0.3 fl oz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatQuantity(tt.quantity, tt.unitType)
			if got != tt.want {
				t.Errorf("formatQuantity(%v, %v) = %q, want %q",
					tt.quantity, tt.unitType, got, tt.want)
			}
		})
	}
}

func TestFormatQuantity_Boundaries(t *testing.T) {
	// Test boundary cases more precisely
	tests := []struct {
		name     string
		quantity float64
		unitType domain.UnitType
		want     string
	}{
		{
			name:     "grams - just under 16 oz threshold",
			quantity: 450.0, // ~15.87 oz
			unitType: domain.UnitTypeGrams,
			want:     "15.9 oz",
		},
		{
			name:     "grams - just over 16 oz threshold",
			quantity: 460.0, // ~16.23 oz
			unitType: domain.UnitTypeGrams,
			want:     "1.0 lbs",
		},
		{
			name:     "liters - just under 1 gallon threshold",
			quantity: 3.7, // ~125 fl oz
			unitType: domain.UnitTypeLiters,
			want:     "125.1 fl oz",
		},
		{
			name:     "liters - just over 1 gallon threshold",
			quantity: 3.9, // ~131.9 fl oz
			unitType: domain.UnitTypeLiters,
			want:     "1.0 gal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatQuantity(tt.quantity, tt.unitType)
			if got != tt.want {
				t.Errorf("formatQuantity(%v, %v) = %q, want %q",
					tt.quantity, tt.unitType, got, tt.want)
			}
		})
	}
}

func TestFormatQuantity_RealWorldExamples(t *testing.T) {
	// Test with realistic inventory values
	tests := []struct {
		name     string
		quantity float64
		unitType domain.UnitType
		want     string
	}{
		{
			name:     "rice bag - 5kg",
			quantity: 5000.0,
			unitType: domain.UnitTypeGrams,
			want:     "11.0 lbs",
		},
		{
			name:     "pasta - 500g",
			quantity: 500.0,
			unitType: domain.UnitTypeGrams,
			want:     "1.1 lbs",
		},
		{
			name:     "spice - 50g",
			quantity: 50.0,
			unitType: domain.UnitTypeGrams,
			want:     "1.8 oz",
		},
		{
			name:     "cereal boxes - 3 count",
			quantity: 3.0,
			unitType: domain.UnitTypeCount,
			want:     "3",
		},
		{
			name:     "water bottles - 12 count",
			quantity: 12.0,
			unitType: domain.UnitTypeCount,
			want:     "12",
		},
		{
			name:     "orange juice - 1.5 liters",
			quantity: 1.5,
			unitType: domain.UnitTypeLiters,
			want:     "50.7 fl oz",
		},
		{
			name:     "milk - just over 1 gallon",
			quantity: 3.8, // 128.5 fl oz, converts to gal
			unitType: domain.UnitTypeLiters,
			want:     "1.0 gal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatQuantity(tt.quantity, tt.unitType)
			if got != tt.want {
				t.Errorf("formatQuantity(%v, %v) = %q, want %q",
					tt.quantity, tt.unitType, got, tt.want)
			}
		})
	}
}
