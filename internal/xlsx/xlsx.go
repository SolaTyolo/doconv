// Package xlsx parses SpreadsheetML (XLSX) into the unified document model.
package xlsx

import (
	"fmt"
	"io"
	"strings"

	"github.com/postship/doconv/internal/model"
	"github.com/postship/doconv/internal/stats"
	"github.com/xuri/excelize/v2"
)

// ParseFile reads an .xlsx workbook from disk.
func ParseFile(path string) (*model.Document, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("xlsx: %w", err)
	}
	defer func() { _ = f.Close() }()
	return buildDocument(f)
}

// ParseReader reads XLSX from a seekable reader (excelize requirement).
func ParseReader(r io.Reader) (*model.Document, error) {
	f, err := excelize.OpenReader(r, excelize.Options{})
	if err != nil {
		return nil, fmt.Errorf("xlsx: %w", err)
	}
	defer func() { _ = f.Close() }()
	return buildDocument(f)
}

func buildDocument(f *excelize.File) (*model.Document, error) {
	doc := &model.Document{
		Format:   model.FormatXlsx,
		Metadata: model.Metadata{},
	}
	if props, err := f.GetDocProps(); err == nil && props != nil {
		doc.Metadata.Title = props.Title
		doc.Metadata.Author = props.Creator
		doc.Metadata.Subject = props.Subject
	}
	sheets := f.GetSheetList()
	doc.Metadata.SheetHint = len(sheets)
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, err
		}
		tbl := rowsToTable(rows)
		sec := model.Section{Name: sheet}
		if len(tbl.Rows) > 0 {
			sec.Elements = append(sec.Elements, model.Element{Kind: model.KindTable, Table: tbl})
		}
		doc.Sections = append(doc.Sections, sec)
	}
	stats.Apply(doc)
	return doc, nil
}

func rowsToTable(rows [][]string) *model.Table {
	t := &model.Table{}
	for _, row := range rows {
		var tr []model.TableCell
		for _, cell := range row {
			txt := strings.TrimSpace(cell)
			p := model.Paragraph{Runs: []model.Run{{Text: txt}}}
			tr = append(tr, model.TableCell{Paragraphs: []model.Paragraph{p}})
		}
		t.Rows = append(t.Rows, tr)
	}
	return t
}
