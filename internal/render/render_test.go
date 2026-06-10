package render

import (
	"strings"
	"testing"

	"github.com/SolaTyolo/doconv/internal/model"
)

func TestToMarkdown_HeadingAndSection(t *testing.T) {
	doc := &model.Document{
		Format: model.FormatDocx,
		Sections: []model.Section{
			{
				Name: "S1",
				Elements: []model.Element{
					{
						Kind: model.KindParagraph,
						Paragraph: &model.Paragraph{
							OutlineLevel: 2,
							Runs:         []model.Run{{Text: "Sub"}},
						},
					},
				},
			},
		},
	}
	out := ToMarkdown(doc, DefaultMarkdownOptions())
	if !strings.Contains(out, "## S1") {
		t.Fatalf("expected section heading: %s", out)
	}
	if !strings.Contains(out, "## Sub") {
		t.Fatalf("expected outline heading: %s", out)
	}
}

func TestToJSON_ContainsFormat(t *testing.T) {
	doc := &model.Document{Format: model.FormatXlsx, Metadata: model.Metadata{Title: "T"}}
	b, err := ToJSON(doc, JSONPretty)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"format"`) {
		t.Fatalf("%s", b)
	}
}
