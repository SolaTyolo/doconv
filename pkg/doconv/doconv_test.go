package doconv

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseReader_XLSX_ToMarkdown(t *testing.T) {
	f := excelize.NewFile()
	if err := f.SetCellValue("Sheet1", "A1", "Hello"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := ParseReader(bytes.NewReader(buf.Bytes()), "t.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	md := ToMarkdown(doc, DefaultMarkdownOptions())
	if !strings.Contains(md, "Hello") {
		t.Fatalf("markdown: %s", md)
	}
}

func TestDetectFromPath(t *testing.T) {
	ft, err := DetectFromPath("a.pptx")
	if err != nil || ft != FormatPptx {
		t.Fatalf("%v %v", ft, err)
	}
}
