package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeStringArray(t *testing.T) {
	t.Run("should return empty array when input is empty", func(t *testing.T) {
		result := SanitizeStringArray([]string{})
		assert.Empty(t, result)
		assert.Equal(t, []string{}, result)
	})

	t.Run("should return empty array when input is nil", func(t *testing.T) {
		result := SanitizeStringArray(nil)
		assert.Empty(t, result)
		assert.Equal(t, []string{}, result)
	})

	t.Run("should normalize single keyword to lowercase and trim spaces", func(t *testing.T) {
		input := []string{"  KEYWORD  "}
		expected := []string{"keyword"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should remove duplicates and normalize", func(t *testing.T) {
		input := []string{"Keyword", "KEYWORD", "keyword", "  keyword  "}
		expected := []string{"keyword"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should filter out empty strings and whitespace-only strings", func(t *testing.T) {
		input := []string{"valid", "", "  ", "\t", "\n", "another"}
		expected := []string{"valid", "another"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should preserve order of first occurrence", func(t *testing.T) {
		input := []string{"first", "second", "FIRST", "third", "SECOND"}
		expected := []string{"first", "second", "third"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle mixed case and spacing", func(t *testing.T) {
		input := []string{"  Go  ", "GOLANG", "go", "Python", "  PYTHON  ", "javascript"}
		expected := []string{"go", "golang", "python", "javascript"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle special characters", func(t *testing.T) {
		input := []string{"C++", "C#", "Node.js", "c++", "C#"}
		expected := []string{"c++", "c#", "node.js"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle unicode characters", func(t *testing.T) {
		input := []string{"Café", "CAFÉ", "naïve", "NAÏVE"}
		expected := []string{"café", "naïve"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle array with only empty/whitespace strings", func(t *testing.T) {
		input := []string{"", "  ", "\t", "\n", "   \t\n   "}
		result := SanitizeStringArray(input)
		assert.Empty(t, result)
		// The function returns nil when no valid elements are found
		assert.Nil(t, result)
	})

	t.Run("should handle large array with many duplicates", func(t *testing.T) {
		input := make([]string, 1000)
		for i := range 1000 {
			switch i % 3 {
			case 0:
				input[i] = "keyword1"
			case 1:
				input[i] = "KEYWORD2"
			default:
				input[i] = "  Keyword3  "
			}
		}
		expected := []string{"keyword1", "keyword2", "keyword3"}
		result := SanitizeStringArray(input)
		assert.Equal(t, expected, result)
	})
}

// Benchmark tests
func BenchmarkSanitizeStringArray(b *testing.B) {
	input := []string{"Go", "GOLANG", "go", "Python", "PYTHON", "javascript", "JavaScript", "C++", "c++"}

	for b.Loop() {
		SanitizeStringArray(input)
	}
}
// go test ./... -bench=. -benchmem -run=^$
func BenchmarkSanitizeStringArrayLarge(b *testing.B) {
	input := make([]string, 1000)
	for i := range 1000 {
		switch i % 5 {
		case 0:
			input[i] = "keyword1"
		case 1:
			input[i] = "KEYWORD2"
		case 2:
			input[i] = "  Keyword3  "
		case 3:
			input[i] = "KEYWORD4"
		default:
			input[i] = "keyword5"
		}
	}

	for b.Loop() {
		SanitizeStringArray(input)
	}
}
