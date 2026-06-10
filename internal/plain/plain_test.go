package plain

import (
	"strings"
	"testing"

	"github.com/SolaTyolo/doconv/internal/render"
)

func TestParseCSV(t *testing.T) {
	doc, err := ParseCSV([]byte("a,b\n1,2\n"), "data.csv")
	if err != nil {
		t.Fatal(err)
	}
	md := render.ToMarkdown(doc, render.DefaultMarkdownOptions())
	if !strings.Contains(md, "a") || !strings.Contains(md, "1") {
		t.Fatalf("markdown: %q", md)
	}
}

func TestParseJSON(t *testing.T) {
	doc, err := ParseJSON([]byte(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != "json" {
		t.Fatalf("format=%q", doc.Format)
	}
}
