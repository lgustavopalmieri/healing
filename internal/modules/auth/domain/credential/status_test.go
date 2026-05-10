package credential_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
)

func TestStatus_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    credential.Status
		expected bool
	}{
		{name: "true para pending", input: credential.StatusPending, expected: true},
		{name: "true para active", input: credential.StatusActive, expected: true},
		{name: "true para locked", input: credential.StatusLocked, expected: true},
		{name: "true para deleted", input: credential.StatusDeleted, expected: true},
		{name: "false para string desconhecida", input: credential.Status("unknown"), expected: false},
		{name: "false para string vazia", input: credential.Status(""), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Valid())
		})
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     credential.Status
		to       credential.Status
		expected bool
	}{
		{name: "pending -> active", from: credential.StatusPending, to: credential.StatusActive, expected: true},
		{name: "pending -> deleted", from: credential.StatusPending, to: credential.StatusDeleted, expected: true},
		{name: "pending -> locked", from: credential.StatusPending, to: credential.StatusLocked, expected: false},
		{name: "pending -> pending", from: credential.StatusPending, to: credential.StatusPending, expected: false},

		{name: "active -> locked", from: credential.StatusActive, to: credential.StatusLocked, expected: true},
		{name: "active -> deleted", from: credential.StatusActive, to: credential.StatusDeleted, expected: true},
		{name: "active -> pending", from: credential.StatusActive, to: credential.StatusPending, expected: false},
		{name: "active -> active", from: credential.StatusActive, to: credential.StatusActive, expected: false},

		{name: "locked -> active", from: credential.StatusLocked, to: credential.StatusActive, expected: true},
		{name: "locked -> deleted", from: credential.StatusLocked, to: credential.StatusDeleted, expected: true},
		{name: "locked -> pending", from: credential.StatusLocked, to: credential.StatusPending, expected: false},
		{name: "locked -> locked", from: credential.StatusLocked, to: credential.StatusLocked, expected: false},

		{name: "deleted -> active", from: credential.StatusDeleted, to: credential.StatusActive, expected: false},
		{name: "deleted -> pending", from: credential.StatusDeleted, to: credential.StatusPending, expected: false},
		{name: "deleted -> locked", from: credential.StatusDeleted, to: credential.StatusLocked, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}
