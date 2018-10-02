package lib_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m-mizutani/AlertResponder/lib"
)

func TestSectionSerialization(t *testing.T) {
	reportID := lib.NewReportID()
	section := lib.Section{}
	section.Text = []string{"a", "b"}
	section.Title = "test1"

	reportData := lib.NewReportData(reportID)
	reportData.SetSection(section)

	// change original data
	section.Text = []string{"c", "b"}
	section.Title = "test2"

	s2 := reportData.Section()
	assert.Equal(t, s2.Title, "test1")
	assert.ElementsMatch(t, s2.Text, []string{"a", "b"})
}
