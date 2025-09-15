package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		color    int
		expected int
	}{
		{"ColorBlue", ColorBlue, 0x3498db},
		{"ColorPurple", ColorPurple, 0x9b59b6},
		{"ColorOrange", ColorOrange, 0xf39c12},
		{"ColorGreen", ColorGreen, 0x2ecc71},
		{"ColorRed", ColorRed, 0xe74c3c},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.color)
		})
	}
}

func TestColorConstantsAreValid(t *testing.T) {
	// Ensure all color constants are valid hex values within Discord's range
	colors := []int{ColorBlue, ColorPurple, ColorOrange, ColorGreen, ColorRed}

	for _, color := range colors {
		assert.GreaterOrEqual(t, color, 0x000000, "Color should be >= 0x000000")
		assert.LessOrEqual(t, color, 0xffffff, "Color should be <= 0xffffff")
	}
}

func TestColorConstantsAreUnique(t *testing.T) {
	colors := []int{ColorBlue, ColorPurple, ColorOrange, ColorGreen, ColorRed}
	colorMap := make(map[int]bool)

	for _, color := range colors {
		assert.False(t, colorMap[color], "Color %x should be unique", color)
		colorMap[color] = true
	}
}