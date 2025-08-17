package main

import (
	"testing"
)

func TestCalculator_Add(t *testing.T) {
	calculator := NewCalculator()

	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 10, 20, 30},
		{"negative numbers", -5, -3, -8},
		{"mixed numbers", 15, -5, 10},
		{"zero", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Multiply(t *testing.T) {
	calculator := NewCalculator()

	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 4, 5, 20},
		{"negative numbers", -3, -4, 12},
		{"mixed numbers", 6, -2, -12},
		{"zero multiplication", 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Divide(t *testing.T) {
	calculator := NewCalculator()

	tests := []struct {
		name      string
		a, b      int
		expected  int
		expectErr bool
	}{
		{"normal division", 20, 4, 5, false},
		{"division by zero", 10, 0, 0, true},
		{"negative result", -15, 3, -5, false},
		{"zero dividend", 0, 5, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.Divide(tt.a, tt.b)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Divide(%d, %d) expected error but got none", tt.a, tt.b)
				}
				return
			}

			if err != nil {
				t.Errorf("Divide(%d, %d) unexpected error: %v", tt.a, tt.b, err)
				return
			}

			if result != tt.expected {
				t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Benchmark tests to demonstrate benchmark configuration in .mage.yaml
func BenchmarkCalculator_Add(b *testing.B) {
	calculator := NewCalculator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.Add(100, 200)
	}
}

func BenchmarkCalculator_Multiply(b *testing.B) {
	calculator := NewCalculator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.Multiply(50, 75)
	}
}

// Integration test example (would be skipped with -short flag)
func TestIntegration_CalculatorWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	calculator := NewCalculator()

	// Simulate a complex workflow
	step1 := calculator.Add(10, 5)
	step2 := calculator.Multiply(step1, 2)
	step3, err := calculator.Divide(step2, 5)
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	expected := 6 // ((10+5)*2)/5 = 6
	if step3 != expected {
		t.Errorf("Integration workflow result = %d; want %d", step3, expected)
	}
}
