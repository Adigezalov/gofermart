package utils

import (
	"testing"
)

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		// Валидные номера
		{"Valid credit card", "4532015112830366", true},
		{"Valid simple", "79927398713", true},
		{"Valid with spaces", "4532 0151 1283 0366", true},
		{"Valid short", "18", true},
		{"Valid from spec 1", "12345678903", true},
		{"Valid from spec 2", "9278923470", true},

		// Невалидные номера
		{"Invalid credit card", "4532015112830367", false},
		{"Invalid simple", "79927398714", false},
		{"Empty string", "", false},
		{"Single digit", "5", false},
		{"Contains letters", "4532a15112830366", false},
		{"Contains special chars", "4532-0151-1283-0366", false},
		{"Only spaces", "   ", false},
		{"Invalid from spec 1", "12345678904", false},
		{"Invalid from spec 2", "9278923471", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLuhn(tt.number)
			if result != tt.expected {
				t.Errorf("IsValidLuhn(%q) = %v, expected %v", tt.number, result, tt.expected)
			}
		})
	}
}

func TestGenerateValidLuhn(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{"Simple prefix", "123456789", "1234567897"},
		{"Short prefix", "1", "18"},
		{"Zero prefix", "0", "00"},
		{"Long prefix", "453201511283036", "4532015112830366"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateValidLuhn(tt.prefix)
			if result != tt.expected {
				t.Errorf("GenerateValidLuhn(%q) = %q, expected %q", tt.prefix, result, tt.expected)
			}

			// Проверяем, что сгенерированный номер действительно валиден
			if !IsValidLuhn(result) {
				t.Errorf("Generated number %q is not valid according to Luhn", result)
			}
		})
	}
}
