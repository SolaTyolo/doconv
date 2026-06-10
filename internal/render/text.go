// Package render converts the unified model to plain text, Markdown, or JSON.
package render

import (
	"strings"

	"github.com/SolaTyolo/doconv/internal/model"
)

// TextOptions configures plain-text rendering.
type TextOptions struct {
	SectionSeparator string // between sections (default: two newlines)
}

// DefaultTextOptions returns sensible defaults.
func DefaultTextOptions() TextOptions {
	return TextOptions{SectionSeparator: "\n\n"}
}

// ToPlainText flattens the document to UTF-8 plain text.
func ToPlainText(doc *model.Document, opt TextOptions) string {
	if doc == nil {
		return ""
	}
	sep := opt.SectionSeparator
	if sep == "" {
		sep = "\n\n"
	}
	var b strings.Builder
	for si, sec := range doc.Sections {
		if si > 0 {
			b.WriteString(sep)
		}
		if sec.Name != "" {
			b.WriteString(sec.Name)
			b.WriteByte('\n')
		}
		b.WriteString(sectionPlainText(&sec))
	}
	return strings.TrimSpace(b.String())
}

func sectionPlainText(sec *model.Section) string {
	var b strings.Builder
	for _, el := range sec.Elements {
		switch el.Kind {
		case model.KindParagraph:
			if el.Paragraph != nil {
				b.WriteString(paragraphPlainText(el.Paragraph))
				b.WriteByte('\n')
			}
		case model.KindTable:
			if el.Table != nil {
				b.WriteString(tablePlainText(el.Table))
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

func paragraphPlainText(p *model.Paragraph) string {
	var b strings.Builder
	for _, r := range p.Runs {
		b.WriteString(r.Text)
	}
	return b.String()
}

func tablePlainText(t *model.Table) string {
	var rows []string
	for _, row := range t.Rows {
		var cells []string
		for _, c := range row {
			var sb strings.Builder
			for _, para := range c.Paragraphs {
				sb.WriteString(paragraphPlainText(&para))
				sb.WriteByte(' ')
			}
			cells = append(cells, strings.TrimSpace(sb.String()))
		}
		rows = append(rows, strings.Join(cells, "\t"))
	}
	return strings.Join(rows, "\n")
}
