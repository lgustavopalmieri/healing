package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/token"
)

func TestHash_SameInputProducesSameOutput(t *testing.T) {
	got1 := token.Hash("refresh-token-value")
	got2 := token.Hash("refresh-token-value")

	assert.Equal(t, got1, got2)
}

func TestHash_DifferentInputsProduceDifferentOutputs(t *testing.T) {
	got1 := token.Hash("a")
	got2 := token.Hash("b")

	assert.NotEqual(t, got1, got2)
}

func TestHash_OutputIsHex64Chars(t *testing.T) {
	got := token.Hash("value")

	assert.Len(t, got, 64)
}

func TestHash_EmptyString(t *testing.T) {
	got := token.Hash("")

	assert.Len(t, got, 64)
}
