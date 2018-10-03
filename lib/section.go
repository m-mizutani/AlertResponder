package lib

import (
	"fmt"
	"strings"
)

type Section struct {
	Title      string
	paragraphs []paragraph
}

// NewSection is a constructor of Section
func NewSection(title string) Section {
	s := Section{
		Title:      title,
		paragraphs: []paragraph{},
	}
	return s
}

// Append adds a paragraph to the Section
func (x *Section) Append(p paragraph) {
	x.paragraphs = append(x.paragraphs, p)
}

func (x *Section) MarkDown() []string {
	s := []string{fmt.Sprintf("### %s", x.Title), ""}
	for _, p := range x.paragraphs {
		s = append(s, p.toMarkDown()...)
		s = append(s, "")
	}
	return s
}

type paragraph interface {
	toMarkDown() []string
}

// ----------------------------------------

// List is a one indent list structure
type List struct {
	items []string
}

// NewList is a constructor of List
func NewList() List {
	return List{items: []string{}}
}

// Append adds line of list
func (x *List) Append(item string) {
	x.items = append(x.items, item)
}

func (x *List) toMarkDown() []string {
	s := []string{}
	for _, item := range x.items {
		s = append(s, fmt.Sprintf("- %s", item))
	}
	return s
}

// ----------------------------------------

// Table is to show table structure of MarkDown
type Table struct {
	Head Row
	Rows []Row
}

// NewTable is a constructor of Table
func NewTable() Table {
	t := Table{}
	return t
}

func (x *Table) toMarkDown() []string {
	// Head
	s := []string{x.Head.toMarkDown()}

	// Border between head and body
	border := []string{""}
	for i := 0; i < len(x.Head.items); i++ {
		border = append(border, ":---------")
	}
	border = append(border, "")
	line := strings.Join(border, "|")
	s = append(s, line)

	// Body
	for _, r := range x.Rows {
		s = append(s, r.toMarkDown())
	}

	return s
}

// Append adds row to Table
func (x *Table) Append(row Row) {
	x.Rows = append(x.Rows, row)
}

type Row struct {
	items []string
}

func NewRow() Row {
	return Row{}
}

func (x *Row) AddItem(item string) {
	x.items = append(x.items, item)
}

func (x *Row) toMarkDown() string {
	content := strings.Join(x.items, " | ")
	return fmt.Sprintf("| %s |", content)
}
