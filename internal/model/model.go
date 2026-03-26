// Package model defines the unified document structure produced by format-specific parsers.
package model

import "time"

// Format identifies the source Office Open XML format.
type Format string

const (
	FormatDocx Format = "docx"
	FormatXlsx Format = "xlsx"
	FormatPptx Format = "pptx"
)

// Document is the normalized representation shared by DOCX, XLSX, and PPTX parsers.
type Document struct {
	Format    Format     `json:"format"`
	Metadata  Metadata   `json:"metadata"`
	Sections  []Section  `json:"sections"`
	Resources Resources  `json:"resources,omitempty"`
	Stats     Statistics `json:"statistics,omitempty"`
}

// Metadata holds common document properties when available.
type Metadata struct {
	Title     string    `json:"title,omitempty"`
	Author    string    `json:"author,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Modified  time.Time `json:"modified,omitempty"`
	Subject   string    `json:"subject,omitempty"`
	SheetHint int       `json:"sheet_count,omitempty"`   // xlsx
	SlideHint int       `json:"slide_count,omitempty"`   // pptx
	SectionCt int       `json:"section_count,omitempty"` // docx rough count
}

// Section groups content (e.g. Word section, Excel sheet, PowerPoint slide).
type Section struct {
	Name     string    `json:"name,omitempty"`
	Elements []Element `json:"elements"`
}

// Element is a paragraph, table, or raw block.
type Element struct {
	Kind      ElementKind `json:"kind"`
	Paragraph *Paragraph  `json:"paragraph,omitempty"`
	Table     *Table      `json:"table,omitempty"`
}

// ElementKind discriminates JSON and internal rendering.
type ElementKind string

const (
	KindParagraph ElementKind = "paragraph"
	KindTable     ElementKind = "table"
)

// Paragraph is a sequence of runs with optional outline level for headings.
type Paragraph struct {
	Runs         []Run `json:"runs"`
	OutlineLevel int   `json:"outline_level,omitempty"` // 0 = body; 1–6 = heading level hint
}

// Run is inline text with basic formatting.
type Run struct {
	Text          string `json:"text"`
	Bold          bool   `json:"bold,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Superscript   bool   `json:"superscript,omitempty"`
	Subscript     bool   `json:"subscript,omitempty"`
	Hyperlink     string `json:"hyperlink,omitempty"`
}

// Table is a grid of cells (each cell may contain multiple paragraphs).
type Table struct {
	Rows [][]TableCell `json:"rows"`
}

// TableCell holds one or more paragraphs inside a table cell.
type TableCell struct {
	Paragraphs []Paragraph `json:"paragraphs"`
}

// Resources lists embedded binaries keyed by a stable id (e.g. media filename).
type Resources map[string][]byte

// Statistics aggregates simple metrics after parsing.
type Statistics struct {
	Words      int `json:"words"`
	Characters int `json:"characters"`
}
