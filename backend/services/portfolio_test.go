package services

import "testing"

func TestSafePercent(t *testing.T) {
	t.Run("zero denominator returns zero", func(t *testing.T) {
		if got := safePercent(10, 0); got != 0 {
			t.Fatalf("expected 0, got %v", got)
		}
	})

	t.Run("valid denominator returns percentage", func(t *testing.T) {
		if got := safePercent(25, 200); got != 12.5 {
			t.Fatalf("expected 12.5, got %v", got)
		}
	})
}
