package lib_test

import (
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/stretchr/testify/assert"
)

func TestSectionSerialization(t *testing.T) {
	reportID := lib.NewReportID()
	section := lib.ReportPage{}
	section.Text = []string{"a", "b"}
	section.Title = "test1"

	reportData := lib.NewReportComponent(reportID)
	reportData.SetPage(section)

	// change original data
	section.Text = []string{"c", "b"}
	section.Title = "test2"

	s2 := reportData.Page()
	assert.Equal(t, s2.Title, "test1")
	assert.ElementsMatch(t, s2.Text, []string{"a", "b"})
}
