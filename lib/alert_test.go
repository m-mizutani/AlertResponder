package lib_test

import (
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/stretchr/testify/assert"
)

func TestAttrMatch(t *testing.T) {
	attr := lib.Attribute{
		Type:    "ipaddr",
		Value:   "10.2.3.4",
		Key:     "source address",
		Context: []string{"local", "server"},
	}

	assert.True(t, attr.Match("local", "ipaddr"))
	assert.True(t, attr.Match("server", "ipaddr"))
	assert.False(t, attr.Match("local", "domain"))
	assert.False(t, attr.Match("remote", "ipaddr"))
}
