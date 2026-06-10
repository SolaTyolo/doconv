// Package plain parses text-first formats into the unified document model.
package plain

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/SolaTyolo/doconv/internal/model"
)

func ParseJSON(data []byte) (*model.Document, error) {
	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		buf.Write(data)
	}
	body := buf.String()
	return model.TextDocument(model.FormatJSON, "JSON", body), nil
}

func ParseCSV(data []byte, filename string) (*model.Document, error) {
	comma := ','
	if strings.HasSuffix(strings.ToLower(filename), ".tsv") {
		comma = '\t'
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	r.Comma = comma
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	if len(rows) == 0 {
		return model.TextDocument(model.FormatCSV, filename, ""), nil
	}

	table := &model.Table{Rows: make([][]model.TableCell, 0, len(rows))}
	for _, row := range rows {
		cells := make([]model.TableCell, len(row))
		for i, cell := range row {
			cells[i] = model.TableCell{
				Paragraphs: []model.Paragraph{{Runs: []model.Run{{Text: cell}}}},
			}
		}
		table.Rows = append(table.Rows, cells)
	}

	return &model.Document{
		Format:   model.FormatCSV,
		Metadata: model.Metadata{Title: filename},
		Sections: []model.Section{{
			Name: filename,
			Elements: []model.Element{{
				Kind:  model.KindTable,
				Table: table,
			}},
		}},
	}, nil
}

func ReadAll(r io.Reader, max int64) ([]byte, error) {
	if max <= 0 {
		max = 32 << 20
	}
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("file exceeds size limit")
	}
	return data, nil
}
