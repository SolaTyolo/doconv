package render

import (
	"bytes"
	"encoding/json"

	"github.com/SolaTyolo/doconv/internal/model"
)

// JSONFormat selects indentation for JSON output.
type JSONFormat string

const (
	JSONPretty  JSONFormat = "pretty"
	JSONCompact JSONFormat = "compact"
)

// ToJSON serializes the full document structure.
func ToJSON(doc *model.Document, jsonFmt JSONFormat) ([]byte, error) {
	if doc == nil {
		return []byte("null"), nil
	}
	if jsonFmt == JSONCompact {
		return json.Marshal(doc)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}
