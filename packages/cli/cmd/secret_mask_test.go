package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "long secret shows last 4", value: "supersecret-api-key-1234", want: "****1234"},
		{name: "exactly 5 chars shows last 4", value: "abcde", want: "****bcde"},
		{name: "4 chars fully masked", value: "abcd", want: "****"},
		{name: "3 chars fully masked", value: "abc", want: "***"},
		{name: "empty string", value: "", want: ""},
		{name: "1 char", value: "x", want: "*"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, maskSecret(tc.value))
		})
	}
}

func TestMaskSecretNeverLeaksFullValue(t *testing.T) {
	// For any secret longer than 4 chars, the masked output must not
	// contain the prefix portion of the secret.
	value := "my-very-long-secret-token-value"
	masked := maskSecret(value)
	assert.True(t, len(masked) <= 4+4, "masked output should be at most 8 chars")
	assert.False(t, strings.Contains(masked, "my-very-long-secret-token"))
	assert.True(t, strings.HasSuffix(masked, "alue")) // last 4 chars of "value"
}

func TestStoreSecretFallbackReturnsError(t *testing.T) {
	err := storeSecretFallback("foo", "bar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no secure keychain available")
}
