package random

import (
	"strings"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	t.Parallel()
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	sizes := []int{1, 5, 10, 15, 20, 25, 30, 35}

	for _, size := range sizes {
		s, err := NewRandomString(size, alphabet)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(s) != size {
			t.Errorf("Expected string length %d, got %d", size, len(s))
		}
		for _, char := range s {
			if !strings.ContainsRune(alphabet, char) {
				t.Errorf("Character %q not found in alphabet", char)
			}
		}
	}
}

func TestNewRandomString_Collisions(t *testing.T) {
	t.Parallel()
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	size := 10
	iterations := 1000000

	seen := make(map[string]struct{}, iterations)
	for i := 0; i < iterations; i++ {
		s, err := NewRandomString(size, alphabet)
		if err != nil {
			t.Fatalf("Error generating random string: %v", err)
		}
		if _, exists := seen[s]; exists {
			t.Fatalf("Collision detected for string: %s; iteration: %d", s, i)
		}
		seen[s] = struct{}{}
	}
}
