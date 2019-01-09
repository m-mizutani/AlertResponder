package lib

import (
	"fmt"
	"strings"
)

// Attribute is element of alert
type Attribute struct {
	Type    string   `json:"type"`
	Value   string   `json:"value"`
	Key     string   `json:"key"`
	Context []string `json:"context"`
}

// TimeRange has timestamps of alert begin and end
type TimeRange struct {
	Init float64 `json:"init"`
	Last float64 `json:"last"`
}

// Alert is extranted data from KinesisStream
type Alert struct {
	Name        string `json:"name"`
	Rule        string `json:"rule"`
	Key         string `json:"key"`
	Description string `json:"description"`

	Timestamp TimeRange   `json:"timestamp"`
	Attrs     []Attribute `json:"attrs"`
}

// Title returns string for Github issue title
func (x *Alert) Title() string {
	return fmt.Sprintf("%s: %s", x.Name, x.Description)
}

// Body returns string for Github issue's main body
func (x *Alert) Body() string {
	lines := []string{
		"## Attributes",
		"",
	}

	for _, attr := range x.Attrs {
		line := fmt.Sprintf("- %s: `%s`", attr.Key, attr.Value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// AddAttribute just appends the attribute to the Alert
func (x *Alert) AddAttribute(attr Attribute) {
	x.Attrs = append(x.Attrs, attr)
}

// AddAttributes appends set of attribute to the Alert
func (x *Alert) AddAttributes(attrs []Attribute) {
	x.Attrs = append(x.Attrs, attrs...)
}

// Match checks attribute type and context.
func (x *Attribute) Match(context, attrType string) bool {
	if x.Type != attrType {
		return false
	}

	for _, ctx := range x.Context {
		if ctx == context {
			return true
		}
	}

	return false
}
