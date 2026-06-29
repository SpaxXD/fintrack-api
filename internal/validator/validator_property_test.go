package validator

import (
	"math"
	"testing"

	"pgregory.net/rapid"
)

// TestProperty17_CentsRoundTrip verifies that for any int64 C > 0,
// converting to decimal and back to cents produces the original value.
//
// **Validates: Requirements 6.1, 6.9**
func TestProperty17_CentsRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a positive int64 value representing cents.
		// Limit upper bound to 10 billion cents ($100M) to stay within float64 exact precision.
		c := rapid.Int64Range(1, 10_000_000_000).Draw(t, "cents")

		decimal := FromCents(c)
		result, err := ToCents(decimal)
		if err != nil {
			t.Fatalf("ToCents(FromCents(%d)) returned error: %v", c, err)
		}
		if result != c {
			t.Fatalf("round-trip failed: ToCents(FromCents(%d)) = %d, want %d", c, result, c)
		}
	})
}

// TestProperty17_DecimalRoundTrip verifies that for any float64 D with exactly
// 2 decimal places (D > 0), converting to cents and back produces the original value.
//
// **Validates: Requirements 6.1, 6.9**
func TestProperty17_DecimalRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a positive integer representing hundredths, then divide by 100
		// to get a value with exactly 2 decimal places.
		// Limit to 10 billion to stay within float64 exact precision for 2 decimal places.
		hundredths := rapid.Int64Range(1, 10_000_000_000).Draw(t, "hundredths")
		d := float64(hundredths) / 100.0

		cents, err := ToCents(d)
		if err != nil {
			t.Fatalf("ToCents(%f) returned error: %v", d, err)
		}

		result := FromCents(cents)

		// Compare with tolerance for floating point representation
		if math.Abs(result-d) > 0.001 {
			t.Fatalf("round-trip failed: FromCents(ToCents(%f)) = %f, want %f", d, result, d)
		}
	})
}
