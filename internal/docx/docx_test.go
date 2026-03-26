package docx

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

const minimalDocumentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r><w:t>Hello</w:t></w:r>
      <w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve"> Bold</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

func TestParseReaderAt_Minimal(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(minimalDocumentXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	doc, err := ParseReaderAt(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != "docx" {
		t.Fatalf("format: got %q", doc.Format)
	}
	if len(doc.Sections) != 1 || len(doc.Sections[0].Elements) < 1 {
		t.Fatalf("sections: %+v", doc.Sections)
	}
	el := doc.Sections[0].Elements[0]
	if el.Paragraph == nil {
		t.Fatal("expected paragraph")
	}
	s := ""
	for _, r := range el.Paragraph.Runs {
		s += r.Text
	}
	if !strings.Contains(s, "Hello") || !strings.Contains(s, "Bold") {
		t.Fatalf("text: %q", s)
	}
}
