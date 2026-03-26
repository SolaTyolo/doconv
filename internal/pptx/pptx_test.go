package pptx

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

const minimalSlide = `<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <a:p><a:r><a:t>SlideTitle</a:t></a:r></a:p>
    </p:spTree>
  </p:cSld>
</p:sld>`

func TestParseReaderAt_SlideText(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("ppt/slides/slide1.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(minimalSlide)); err != nil {
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
	if doc.Format != "pptx" {
		t.Fatalf("format %q", doc.Format)
	}
	var sb strings.Builder
	for _, sec := range doc.Sections {
		for _, el := range sec.Elements {
			if el.Paragraph != nil {
				for _, r := range el.Paragraph.Runs {
					sb.WriteString(r.Text)
				}
			}
		}
	}
	if !strings.Contains(sb.String(), "SlideTitle") {
		t.Fatalf("got %q", sb.String())
	}
}
