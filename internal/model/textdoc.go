package model

import "strings"

// TextDocument builds a single-section document from plain text (PDF, JSON, CSV, …).
func TextDocument(format Format, title, body string) *Document {
	body = strings.TrimSpace(body)
	doc := &Document{
		Format: format,
		Metadata: Metadata{
			Title: title,
		},
		Sections: []Section{{
			Name: title,
			Elements: []Element{{
				Kind: KindParagraph,
				Paragraph: &Paragraph{
					Runs: []Run{{Text: body}},
				},
			}},
		}},
	}
	if body != "" {
		doc.Stats.Characters = len([]rune(body))
		doc.Stats.Words = len(strings.Fields(body))
	}
	return doc
}
