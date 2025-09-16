package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnhanvedCode(t *testing.T) {

	ec := EnhancedCode{5, 1, 1}
	assert.Equal(t, "5.1.1", ec.String())
	assert.Equal(t, [3]int{5, 1, 1}, ec.Int())

}
