package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUniques(t *testing.T) {
	slice := []string{"a", "b", "c", "a"}
	expected := []string{"a", "b", "c"}
	assert.Equal(t, expected, Uniques(slice))

	slice = []string{}
	expected = []string(nil)
	assert.Equal(t, expected, Uniques(slice))
}
