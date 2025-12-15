package transformations

import (
	"testing"
)

func TestCreateBorderRadiusFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "valid single value",
			input:    "10",
			hasError: false,
		},
		{
			name:     "valid percentage",
			input:    "25%",
			hasError: false,
		},
		{
			name:     "invalid value",
			input:    "invalid",
			hasError: true,
		},
		{
			name:     "empty value",
			input:    "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := CreateBorderRadiusFilter(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if filter != nil {
					t.Errorf("Expected nil filter on error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if filter == nil {
				t.Errorf("Expected filter but got nil")
			}
		})
	}
}
