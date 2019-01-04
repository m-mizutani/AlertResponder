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

func TestAddAttribute(t *testing.T) {
	attr := lib.Attribute{
		Type:    "ipaddr",
		Value:   "10.2.3.4",
		Key:     "source address",
		Context: []string{"local", "server"},
	}

	alert := lib.Alert{}
	alert.AddAttribute(attr)
	assert.Equal(t, 1, len(alert.Attrs))
	assert.Equal(t, "10.2.3.4", alert.Attrs[0].Value)
}
