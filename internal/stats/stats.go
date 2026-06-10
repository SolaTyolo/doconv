// Package stats computes word and character counts from the unified model.
package stats

import (
	"strings"
	"unicode"

	"github.com/SolaTyolo/doconv/internal/model"
)

// Apply sets document.Statistics from its content.
func Apply(doc *model.Document) {
	var words, chars int
	for _, sec := range doc.Sections {
		for _, el := range sec.Elements {
			switch el.Kind {
			case model.KindParagraph:
				if el.Paragraph != nil {
					w, c := countParagraph(el.Paragraph)
					words += w
					chars += c
				}
			case model.KindTable:
				if el.Table != nil {
					for _, row := range el.Table.Rows {
						for _, cell := range row {
							for _, p := range cell.Paragraphs {
								w, c := countParagraph(&p)
								words += w
								chars += c
							}
						}
					}
				}
			}
		}
	}
	doc.Stats.Words = words
	doc.Stats.Characters = chars
}

func countParagraph(p *model.Paragraph) (words, chars int) {
	var b strings.Builder
	for _, r := range p.Runs {
		b.WriteString(r.Text)
	}
	s := b.String()
	chars = len([]rune(s))
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	for _, f := range fields {
		if len(strings.TrimSpace(f)) > 0 {
			words++
		}
	}
	return words, chars
}
