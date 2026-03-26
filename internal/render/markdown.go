package render

import (
	"strings"

	"github.com/postship/doconv/internal/model"
)

// MarkdownOptions configures Markdown output.
type MarkdownOptions struct {
	Frontmatter      bool
	MaxHeadingLevel  int // clamp 1–6; 0 means no clamp
	TableMode        TableMode
	SectionAsHeading bool // emit section name as ## when true
}

// TableMode selects how tables are serialized when Markdown is insufficient.
type TableMode string

const (
	TableMarkdown TableMode = "markdown"
	TableHTML     TableMode = "html"
	TableASCII    TableMode = "ascii"
)

// DefaultMarkdownOptions returns defaults aligned with common CLI expectations.
func DefaultMarkdownOptions() MarkdownOptions {
	return MarkdownOptions{
		Frontmatter:      false,
		MaxHeadingLevel:  6,
		TableMode:        TableMarkdown,
		SectionAsHeading: true,
	}
}

// ToMarkdown renders the document as GitHub-flavored Markdown.
func ToMarkdown(doc *model.Document, opt MarkdownOptions) string {
	if doc == nil {
		return ""
	}
	var b strings.Builder
	if opt.Frontmatter {
		b.WriteString("---\n")
		if doc.Metadata.Title != "" {
			b.WriteString("title: ")
			b.WriteString(escapeYAMLString(doc.Metadata.Title))
			b.WriteByte('\n')
		}
		if doc.Metadata.Author != "" {
			b.WriteString("author: ")
			b.WriteString(escapeYAMLString(doc.Metadata.Author))
			b.WriteByte('\n')
		}
		if !doc.Metadata.Created.IsZero() {
			b.WriteString("created: ")
			b.WriteString(doc.Metadata.Created.UTC().Format("2006-01-02T15:04:05Z07:00"))
			b.WriteByte('\n')
		}
		b.WriteString("---\n\n")
	}
	for _, sec := range doc.Sections {
		if opt.SectionAsHeading && sec.Name != "" {
			b.WriteString("## ")
			b.WriteString(sec.Name)
			b.WriteString("\n\n")
		}
		b.WriteString(sectionMarkdown(&sec, opt))
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func escapeYAMLString(s string) string {
	if strings.ContainsAny(s, `:"'\n`) {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func sectionMarkdown(sec *model.Section, opt MarkdownOptions) string {
	var b strings.Builder
	for _, el := range sec.Elements {
		switch el.Kind {
		case model.KindParagraph:
			if el.Paragraph != nil {
				b.WriteString(paragraphMarkdown(el.Paragraph, opt))
				b.WriteString("\n\n")
			}
		case model.KindTable:
			if el.Table != nil {
				b.WriteString(tableMarkdown(el.Table, opt))
				b.WriteString("\n\n")
			}
		}
	}
	return b.String()
}

func paragraphMarkdown(p *model.Paragraph, opt MarkdownOptions) string {
	lvl := p.OutlineLevel
	if opt.MaxHeadingLevel > 0 && lvl > opt.MaxHeadingLevel {
		lvl = opt.MaxHeadingLevel
	}
	if lvl >= 1 && lvl <= 6 {
		prefix := strings.Repeat("#", lvl) + " "
		return prefix + runsInlinePlain(p.Runs)
	}
	return runsInlineMarkdown(p.Runs)
}

func runsInlinePlain(runs []model.Run) string {
	var b strings.Builder
	for _, r := range runs {
		b.WriteString(r.Text)
	}
	return b.String()
}

func runsInlineMarkdown(runs []model.Run) string {
	var b strings.Builder
	for _, r := range runs {
		t := escapeMarkdownText(r.Text)
		switch {
		case r.Hyperlink != "" && t != "":
			b.WriteString("[")
			b.WriteString(t)
			b.WriteString("](")
			b.WriteString(escapeMarkdownLinkDest(r.Hyperlink))
			b.WriteString(")")
		case r.Bold && r.Italic:
			b.WriteString("***")
			b.WriteString(t)
			b.WriteString("***")
		case r.Bold:
			b.WriteString("**")
			b.WriteString(t)
			b.WriteString("**")
		case r.Italic:
			b.WriteString("*")
			b.WriteString(t)
			b.WriteString("*")
		case r.Strikethrough:
			b.WriteString("~~")
			b.WriteString(t)
			b.WriteString("~~")
		default:
			b.WriteString(t)
		}
	}
	return b.String()
}

func escapeMarkdownText(s string) string {
	return strings.ReplaceAll(s, "\n", "  \n")
}

func escapeMarkdownLinkDest(s string) string {
	return strings.ReplaceAll(s, ")", "\\)")
}

func tableMarkdown(t *model.Table, opt MarkdownOptions) string {
	if len(t.Rows) == 0 {
		return ""
	}
	switch opt.TableMode {
	case TableHTML:
		return tableHTML(t)
	case TableASCII:
		return tableASCII(t)
	default:
		return tableMarkdownGrid(t)
	}
}

func tableMarkdownGrid(t *model.Table) string {
	var b strings.Builder
	for i, row := range t.Rows {
		var cells []string
		for _, c := range row {
			var sb strings.Builder
			for _, para := range c.Paragraphs {
				sb.WriteString(runsInlineMarkdown(para.Runs))
			}
			cells = append(cells, strings.TrimSpace(sb.String()))
		}
		b.WriteString("| ")
		b.WriteString(strings.Join(cells, " | "))
		b.WriteString(" |\n")
		if i == 0 {
			b.WriteString("|")
			for range cells {
				b.WriteString(" --- |")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func tableHTML(t *model.Table) string {
	var b strings.Builder
	b.WriteString("<table>\n")
	for _, row := range t.Rows {
		b.WriteString("<tr>")
		for _, c := range row {
			b.WriteString("<td>")
			for _, para := range c.Paragraphs {
				b.WriteString(runsInlineMarkdown(para.Runs))
			}
			b.WriteString("</td>")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</table>")
	return b.String()
}

func tableASCII(t *model.Table) string {
	return tableMarkdownGrid(t)
}
