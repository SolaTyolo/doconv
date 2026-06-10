// Package pptx parses PresentationML (PPTX) into the unified document model.
package pptx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/SolaTyolo/doconv/internal/model"
	"github.com/SolaTyolo/doconv/internal/ooxml"
	"github.com/SolaTyolo/doconv/internal/stats"
)

// ParseFile reads a .pptx file from disk.
func ParseFile(path string) (*model.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return ParseReaderAt(f, st.Size())
}

// ParseReaderAt reads PPTX from a random-access reader.
func ParseReaderAt(r io.ReaderAt, size int64) (*model.Document, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("pptx: open zip: %w", err)
	}
	doc := &model.Document{
		Format:   model.FormatPptx,
		Metadata: model.Metadata{},
		Sections: []model.Section{},
	}
	if raw, err := readZipEntry(zr, "docProps/core.xml"); err == nil {
		ooxml.ApplyCoreXML(raw, &doc.Metadata)
	}

	slideNames := listSlideXMLs(zr)
	doc.Metadata.SlideHint = len(slideNames)
	for _, name := range slideNames {
		raw, err := readZipEntry(zr, name)
		if err != nil {
			return nil, err
		}
		title := slideTitleFromPath(name)
		text := extractSlideText(bytes.NewReader(raw))
		text = strings.TrimSpace(text)
		sec := model.Section{Name: title}
		if text != "" {
			p := &model.Paragraph{Runs: []model.Run{{Text: text}}}
			sec.Elements = append(sec.Elements, model.Element{Kind: model.KindParagraph, Paragraph: p})
		}
		doc.Sections = append(doc.Sections, sec)
	}
	stats.Apply(doc)
	return doc, nil
}

func readZipEntry(zr *zip.Reader, want string) ([]byte, error) {
	for _, f := range zr.File {
		if strings.EqualFold(f.Name, want) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(rc)
			_ = rc.Close()
			return b, err
		}
	}
	return nil, fmt.Errorf("missing %s", want)
}

func listSlideXMLs(zr *zip.Reader) []string {
	var names []string
	prefix := "ppt/slides/slide"
	for _, f := range zr.File {
		n := f.Name
		if !strings.HasPrefix(strings.ToLower(n), strings.ToLower(prefix)) {
			continue
		}
		base := path.Base(n)
		if !strings.HasSuffix(strings.ToLower(base), ".xml") {
			continue
		}
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		return slideIndex(names[i]) < slideIndex(names[j])
	})
	return names
}

func slideIndex(name string) int {
	base := strings.TrimSuffix(path.Base(name), ".xml")
	base = strings.TrimPrefix(strings.ToLower(base), "slide")
	n, _ := strconv.Atoi(base)
	return n
}

func slideTitleFromPath(name string) string {
	idx := slideIndex(name)
	if idx > 0 {
		return fmt.Sprintf("Slide %d", idx)
	}
	return path.Base(name)
}

// extractSlideText collects DrawingML text nodes (a:t) in document order.
func extractSlideText(r io.Reader) string {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var parts []string
	inT := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings.Join(parts, " ")
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				inT = true
			}
		case xml.EndElement:
			if se.Name.Local == "t" {
				inT = false
			}
		case xml.CharData:
			if inT {
				s := strings.TrimSpace(string(se))
				if s != "" {
					parts = append(parts, s)
				}
			}
		}
	}
	return strings.Join(parts, " ")
}
