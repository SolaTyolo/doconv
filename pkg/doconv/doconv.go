// Package doconv extracts text and structure from DOCX, XLSX, and PPTX files
// and renders Markdown, plain text, or JSON. It is inspired by the undoc Rust project.
package doconv

import (
	"io"

	"github.com/postship/doconv/internal/model"
	"github.com/postship/doconv/internal/parse"
	"github.com/postship/doconv/internal/render"
)

// Re-export model types for public API stability.
type (
	Document    = model.Document
	Format      = model.Format
	Metadata    = model.Metadata
	Section     = model.Section
	Element     = model.Element
	ElementKind = model.ElementKind
	Paragraph   = model.Paragraph
	Run         = model.Run
	Table       = model.Table
	TableCell   = model.TableCell
	Statistics  = model.Statistics
)

// Format constants.
const (
	FormatDocx = model.FormatDocx
	FormatXlsx = model.FormatXlsx
	FormatPptx = model.FormatPptx
)

// ParseFile opens a path and returns a unified document tree.
func ParseFile(path string) (*Document, error) {
	return parse.ParseFile(path)
}

// ParseReader reads a full document from r. Use filenameHint like "file.docx" when possible.
func ParseReader(r io.Reader, filenameHint string) (*Document, error) {
	return parse.ParseReader(r, filenameHint)
}

// ParseReaderAt parses from a random-access reader when format is known.
func ParseReaderAt(ra io.ReaderAt, size int64, ft Format) (*Document, error) {
	return parse.ParseReaderAt(ra, size, model.Format(ft))
}

// MarkdownOptions configures Markdown rendering (see internal defaults).
type MarkdownOptions = render.MarkdownOptions

// TableMode selects table serialization.
type TableMode = render.TableMode

// Table mode constants.
const (
	TableMarkdown = render.TableMarkdown
	TableHTML     = render.TableHTML
	TableASCII    = render.TableASCII
)

// TextOptions configures plain-text rendering.
type TextOptions = render.TextOptions

// JSONFormat selects JSON indentation.
type JSONFormat = render.JSONFormat

const (
	JSONPretty  = render.JSONPretty
	JSONCompact = render.JSONCompact
)

// DefaultMarkdownOptions returns default Markdown settings.
func DefaultMarkdownOptions() MarkdownOptions {
	return render.DefaultMarkdownOptions()
}

// DefaultTextOptions returns default plain-text settings.
func DefaultTextOptions() TextOptions {
	return render.DefaultTextOptions()
}

// ToMarkdown converts a parsed document to Markdown.
func ToMarkdown(doc *Document, opt MarkdownOptions) string {
	return render.ToMarkdown(doc, opt)
}

// ToPlainText converts a parsed document to plain text.
func ToPlainText(doc *Document, opt TextOptions) string {
	return render.ToPlainText(doc, opt)
}

// ToJSON converts a parsed document to JSON bytes.
func ToJSON(doc *Document, jsonFmt JSONFormat) ([]byte, error) {
	return render.ToJSON(doc, jsonFmt)
}
