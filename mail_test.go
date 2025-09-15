package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAddress(t *testing.T) {

	tests := []struct {
		address        string
		expectedLocal  string
		expectedDomain string
	}{
		{"test@hyvor.com", "test", "hyvor.com"},
		{"supun+contact@gmail.com", "supun+contact", "gmail.com"},
		{"invalidaddress", "invalidaddress", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		local, domain := splitAddress(tt.address)
		assert.Equal(t, tt.expectedLocal, local)
		assert.Equal(t, tt.expectedDomain, domain)
	}

}
