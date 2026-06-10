// Package docx parses WordprocessingML (DOCX) into the unified document model.
package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/SolaTyolo/doconv/internal/model"
	"github.com/SolaTyolo/doconv/internal/ooxml"
	"github.com/SolaTyolo/doconv/internal/stats"
)

// ParseFile reads a .docx file from disk.
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

// ParseReaderAt reads DOCX from a random-access reader (ZIP requirement).
func ParseReaderAt(r io.ReaderAt, size int64) (*model.Document, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("docx: open zip: %w", err)
	}
	doc := &model.Document{
		Format:   model.FormatDocx,
		Metadata: model.Metadata{},
		Sections: []model.Section{},
	}
	if raw, err := readZipFile(zr, "docProps/core.xml"); err == nil {
		ooxml.ApplyCoreXML(raw, &doc.Metadata)
	}
	raw, err := readZipFile(zr, "word/document.xml")
	if err != nil {
		return nil, fmt.Errorf("docx: %w", err)
	}
	section, err := parseDocumentBody(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	doc.Sections = []model.Section{{Name: "Document", Elements: section}}
	doc.Metadata.SectionCt = len(doc.Sections)
	stats.Apply(doc)
	return doc, nil
}

func readZipFile(zr *zip.Reader, want string) ([]byte, error) {
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

func parseDocumentBody(r io.Reader) ([]model.Element, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "body" {
			return parseBodyChildren(dec, &se)
		}
	}
	return nil, fmt.Errorf("docx: no document body")
}

func parseBodyChildren(dec *xml.Decoder, bodyStart *xml.StartElement) ([]model.Element, error) {
	var out []model.Element
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "p":
				p, err := parseParagraph(dec, &se)
				if err != nil {
					return nil, err
				}
				if p != nil && paragraphHasText(p) {
					out = append(out, model.Element{Kind: model.KindParagraph, Paragraph: p})
				}
			case "tbl":
				tbl, err := parseTable(dec, &se)
				if err != nil {
					return nil, err
				}
				if tbl != nil && len(tbl.Rows) > 0 {
					out = append(out, model.Element{Kind: model.KindTable, Table: tbl})
				}
			default:
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == bodyStart.Name.Local && se.Name.Space == bodyStart.Name.Space {
				return out, nil
			}
		}
	}
}

func paragraphHasText(p *model.Paragraph) bool {
	for _, r := range p.Runs {
		if strings.TrimSpace(r.Text) != "" {
			return true
		}
	}
	return false
}

func parseParagraph(dec *xml.Decoder, start *xml.StartElement) (*model.Paragraph, error) {
	p := &model.Paragraph{Runs: []model.Run{}}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "pPr":
				pp := &paraProps{}
				if err := scanPPr(dec, &se, pp); err != nil {
					return nil, err
				}
				p.OutlineLevel = pp.outline
				if pp.style != "" {
					if lvl := headingLevelFromStyle(pp.style); lvl > 0 {
						p.OutlineLevel = lvl
					}
				}
			case "r":
				runs, err := parseRuns(dec, &se)
				if err != nil {
					return nil, err
				}
				p.Runs = append(p.Runs, runs...)
			case "hyperlink":
				runs, err := parseHyperlink(dec, &se)
				if err != nil {
					return nil, err
				}
				p.Runs = append(p.Runs, runs...)
			default:
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return p, nil
			}
		}
	}
}

type paraProps struct {
	style   string
	outline int
}

func scanPPr(dec *xml.Decoder, start *xml.StartElement, out *paraProps) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "pStyle":
				for _, a := range se.Attr {
					if a.Name.Local == "val" {
						out.style = a.Value
					}
				}
			case "outlineLvl":
				for _, a := range se.Attr {
					if a.Name.Local == "val" {
						if v, err := strconv.Atoi(a.Value); err == nil {
							out.outline = v + 1
						}
					}
				}
				_ = skipSubtree(dec, &se)
			default:
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}

func headingLevelFromStyle(style string) int {
	s := strings.ToLower(style)
	for i := 1; i <= 6; i++ {
		if strings.Contains(s, "heading") && strings.Contains(s, strconv.Itoa(i)) {
			return i
		}
	}
	// Common built-in IDs: Heading1, heading 1, etc.
	if strings.HasPrefix(s, "heading") {
		rest := strings.TrimPrefix(s, "heading")
		rest = strings.TrimSpace(rest)
		if n, err := strconv.Atoi(rest); err == nil && n >= 1 && n <= 6 {
			return n
		}
	}
	return 0
}

func parseRuns(dec *xml.Decoder, start *xml.StartElement) ([]model.Run, error) {
	var runs []model.Run
	state := runState{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "rPr":
				if err := scanRPr(dec, &se, &state); err != nil {
					return nil, err
				}
			case "t":
				var sb strings.Builder
				if err := readTextContent(dec, &se, &sb); err != nil {
					return nil, err
				}
				txt := sb.String()
				if txt != "" {
					r := model.Run{Text: txt, Bold: state.bold, Italic: state.italic,
						Underline: state.underline, Strikethrough: state.strike,
						Superscript: state.super, Subscript: state.sub}
					runs = append(runs, r)
				}
			case "tab":
				runs = append(runs, model.Run{Text: "\t", Bold: state.bold, Italic: state.italic,
					Underline: state.underline, Strikethrough: state.strike,
					Superscript: state.super, Subscript: state.sub})
			case "br":
				runs = append(runs, model.Run{Text: "\n", Bold: state.bold, Italic: state.italic,
					Underline: state.underline, Strikethrough: state.strike,
					Superscript: state.super, Subscript: state.sub})
			default:
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return runs, nil
			}
		}
	}
}

type runState struct {
	bold, italic, underline, strike, super, sub bool
}

func scanRPr(dec *xml.Decoder, start *xml.StartElement, st *runState) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "b", "bCs":
				st.bold = true
			case "i", "iCs":
				st.italic = true
			case "u":
				st.underline = true
			case "strike", "dstrike":
				st.strike = true
			case "vertAlign":
				for _, a := range se.Attr {
					if a.Name.Local == "val" {
						v := strings.ToLower(a.Value)
						if v == "superscript" {
							st.super = true
						}
						if v == "subscript" {
							st.sub = true
						}
					}
				}
			default:
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}

func readTextContent(dec *xml.Decoder, start *xml.StartElement, sb *strings.Builder) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.CharData:
			sb.WriteString(string(t))
		case xml.EndElement:
			if t.Name.Local == start.Name.Local && t.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}

func parseHyperlink(dec *xml.Decoder, start *xml.StartElement) ([]model.Run, error) {
	var href string
	for _, a := range start.Attr {
		if a.Name.Local == "id" || a.Name.Local == "anchor" {
			href = a.Value
			break
		}
	}
	var all []model.Run
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "r" {
				runs, err := parseRuns(dec, &se)
				if err != nil {
					return nil, err
				}
				for i := range runs {
					if href != "" {
						runs[i].Hyperlink = href
					}
				}
				all = append(all, runs...)
			} else {
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return all, nil
			}
		}
	}
}

func parseTable(dec *xml.Decoder, start *xml.StartElement) (*model.Table, error) {
	t := &model.Table{Rows: nil}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "tr" {
				row, err := parseRow(dec, &se)
				if err != nil {
					return nil, err
				}
				t.Rows = append(t.Rows, row)
			} else if se.Name.Local != "tblPr" && se.Name.Local != "tblGrid" {
				_ = skipSubtree(dec, &se)
			} else {
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return t, nil
			}
		}
	}
}

func parseRow(dec *xml.Decoder, start *xml.StartElement) ([]model.TableCell, error) {
	var cells []model.TableCell
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "tc" {
				cell, err := parseCell(dec, &se)
				if err != nil {
					return nil, err
				}
				cells = append(cells, cell)
			} else {
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return cells, nil
			}
		}
	}
}

func parseCell(dec *xml.Decoder, start *xml.StartElement) (model.TableCell, error) {
	var paras []model.Paragraph
	for {
		tok, err := dec.Token()
		if err != nil {
			return model.TableCell{}, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "p" {
				p, err := parseParagraph(dec, &se)
				if err != nil {
					return model.TableCell{}, err
				}
				if p != nil {
					paras = append(paras, *p)
				}
			} else {
				_ = skipSubtree(dec, &se)
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return model.TableCell{Paragraphs: paras}, nil
			}
		}
	}
}

func skipSubtree(dec *xml.Decoder, start *xml.StartElement) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			_ = skipSubtree(dec, &se)
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}
