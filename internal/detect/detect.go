// Package detect identifies Office Open XML formats from paths or magic bytes.
package detect

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/postship/doconv/internal/model"
)

var (
	// ErrUnknownFormat is returned when bytes or extension do not match supported types.
	ErrUnknownFormat = errors.New("unknown or unsupported document format")
)

// FromPath uses file extension (case-insensitive).
func FromPath(path string) (model.Format, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".docx":
		return model.FormatDocx, nil
	case ".xlsx":
		return model.FormatXlsx, nil
	case ".pptx":
		return model.FormatPptx, nil
	default:
		return "", ErrUnknownFormat
	}
}

// ZIP local file header signature PK\x03\x04
var zipLocalHeader = []byte{0x50, 0x4b, 0x03, 0x04}

// FromBytes inspects the ZIP signature and required OOXML parts.
// At least 4 bytes are read; full detection may need more for OOXML checks.
func FromBytes(data []byte) (model.Format, error) {
	if len(data) < 4 || !bytes.HasPrefix(data, zipLocalHeader) {
		return "", ErrUnknownFormat
	}
	// Peek into ZIP: use minimal reader
	r := bytes.NewReader(data)
	zr, err := zip.NewReader(r, int64(len(data)))
	if err != nil {
		return "", ErrUnknownFormat
	}

	has := func(name string) bool {
		for _, f := range zr.File {
			if strings.EqualFold(f.Name, name) {
				return true
			}
		}
		return false
	}

	switch {
	case has("[Content_Types].xml") && has("word/document.xml"):
		return model.FormatDocx, nil
	case has("[Content_Types].xml") && has("xl/workbook.xml"):
		return model.FormatXlsx, nil
	case has("[Content_Types].xml") && has("ppt/presentation.xml"):
		return model.FormatPptx, nil
	default:
		return "", ErrUnknownFormat
	}
}

// FromReader buffers up to maxPeek bytes for detection (default cap in caller).
func FromReader(r io.Reader, maxPeek int) (model.Format, []byte, error) {
	if maxPeek <= 0 {
		maxPeek = 512 * 1024
	}
	buf, err := io.ReadAll(io.LimitReader(r, int64(maxPeek)))
	if err != nil {
		return "", nil, err
	}
	ft, err := FromBytes(buf)
	return ft, buf, err
}
