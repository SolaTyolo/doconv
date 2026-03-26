package xlsx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseReader_SimpleSheet(t *testing.T) {
	f := excelize.NewFile()
	if err := f.SetCellValue("Sheet1", "A1", "Name"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "B1", "Value"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "A2", "x"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := ParseReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != "xlsx" {
		t.Fatalf("format %q", doc.Format)
	}
	if len(doc.Sections) < 1 {
		t.Fatal("expected at least one section")
	}
	md := ""
	for _, sec := range doc.Sections {
		for _, el := range sec.Elements {
			if el.Table != nil {
				for _, row := range el.Table.Rows {
					for _, c := range row {
						for _, p := range c.Paragraphs {
							for _, r := range p.Runs {
								md += r.Text + " "
							}
						}
					}
				}
			}
		}
	}
	if !strings.Contains(md, "Name") || !strings.Contains(md, "Value") {
		t.Fatalf("content: %q", md)
	}
}
