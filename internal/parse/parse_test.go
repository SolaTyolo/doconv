package parse

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseReader_WithHint(t *testing.T) {
	f := excelize.NewFile()
	_ = f.SetCellValue("Sheet1", "A1", "x")
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := ParseReader(bytes.NewReader(buf.Bytes()), "book.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != "xlsx" {
		t.Fatalf("got %q", doc.Format)
	}
}
