package lib_test

import (
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListToMarkDown(t *testing.T) {
	s := lib.NewSection()
	s.Title = "abc"

	l := lib.NewList()
	l.Append("x")
	l.Append("y")
	s.Append(&l)

	lines := s.MarkDown()
	require.Equal(t, 5, len(lines))
	assert.Equal(t, "### abc", lines[0])
	assert.Equal(t, "", lines[1])
	assert.Equal(t, "- x", lines[2])
	assert.Equal(t, "- y", lines[3])
	assert.Equal(t, "", lines[4])
}

func TestTableToMarkDown(t *testing.T) {
	s := lib.NewSection()
	s.Title = "abc"

	tbl := lib.NewTable()
	tbl.Head.AddItem("blue")
	tbl.Head.AddItem("orange")
	tbl.Head.AddItem("magic")

	r1 := lib.NewRow()
	r1.AddItem("five")
	r1.AddItem("timeless")
	r1.AddItem("words")
	r2 := lib.NewRow()
	r2.AddItem("x")
	r2.AddItem("y")
	r2.AddItem("z")

	tbl.Append(r1)
	tbl.Append(r2)

	s.Append(&tbl)

	lines := s.MarkDown()

	require.Equal(t, 7, len(lines))
	assert.Contains(t, lines[0], "abc")
	assert.Equal(t, "", lines[1])
	assert.Contains(t, lines[2], "| blue |")
	assert.Contains(t, lines[2], "| orange |")
	assert.NotContains(t, lines[3], "blue")
	assert.NotContains(t, lines[3], "five")
	assert.Contains(t, lines[4], "| words |")
}
