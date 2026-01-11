---
inclusion: always
---

# Do not write comments inside code.

Examples of disallowed content:
- // explaining logic
- /* block comments */
- function or struct comments

Only include comments if explicitly requested.

# Do NOT write benchmark tests.
Benchmark tests must be created ONLY if explicitly requested by the user.

# All tests MUST:
- Use the testify/assert library
- Use t.Run with descriptive, behavior-driven test names
- Follow the same structure and style as the examples provided by the user.

Example of structure and style:
```go
func TestSanitizeStringArray(t *testing.T) {
	t.Run("should return empty array when input is empty", func(t *testing.T) {
		result := SanitizeStringArray([]string{})
		assert.Empty(t, result)
		assert.Equal(t, []string{}, result)
	})
}
```

# Do NOT use:
- Manual assertions (if/else, t.Errorf, t.Fatal)
- Other assertion libraries