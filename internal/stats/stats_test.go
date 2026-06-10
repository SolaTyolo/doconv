package stats

import (
	"testing"

	"github.com/SolaTyolo/doconv/internal/model"
)

func TestApply(t *testing.T) {
	doc := &model.Document{
		Sections: []model.Section{
			{
				Elements: []model.Element{
					{Kind: model.KindParagraph, Paragraph: &model.Paragraph{
						Runs: []model.Run{{Text: "one two"}},
					}},
				},
			},
		},
	}
	Apply(doc)
	if doc.Stats.Words < 1 || doc.Stats.Characters < 1 {
		t.Fatalf("stats: %+v", doc.Stats)
	}
}
