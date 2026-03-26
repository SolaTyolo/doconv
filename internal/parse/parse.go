// Package parse dispatches to format-specific parsers from paths or byte streams.
package parse

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/postship/doconv/internal/detect"
	"github.com/postship/doconv/internal/docx"
	"github.com/postship/doconv/internal/model"
	"github.com/postship/doconv/internal/pptx"
	"github.com/postship/doconv/internal/xlsx"
)

// ParseFile detects format from extension and parses the file.
func ParseFile(path string) (*model.Document, error) {
	ft, err := detect.FromPath(path)
	if err != nil {
		return nil, err
	}
	switch ft {
	case model.FormatDocx:
		return docx.ParseFile(path)
	case model.FormatXlsx:
		return xlsx.ParseFile(path)
	case model.FormatPptx:
		return pptx.ParseFile(path)
	default:
		return nil, fmt.Errorf("parse: unsupported format %q", ft)
	}
}

// ParseReader reads the entire stream into memory, detects OOXML type from ZIP layout,
// and parses. filenameHint is used for extension-based detection first (e.g. "a.docx").
func ParseReader(r io.Reader, filenameHint string) (*model.Document, error) {
	if ft, err := detect.FromPath(filenameHint); err == nil {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return parseBytesWithFormat(data, ft)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ft, err := detect.FromBytes(data)
	if err != nil {
		return nil, err
	}
	return parseBytesWithFormat(data, ft)
}

func parseBytesWithFormat(data []byte, ft model.Format) (*model.Document, error) {
	br := bytes.NewReader(data)
	switch ft {
	case model.FormatDocx:
		return docx.ParseReaderAt(br, int64(len(data)))
	case model.FormatXlsx:
		return xlsx.ParseReader(br)
	case model.FormatPptx:
		return pptx.ParseReaderAt(br, int64(len(data)))
	default:
		return nil, fmt.Errorf("parse: unsupported format %q", ft)
	}
}

// ParseReaderAt parses from a random-access reader when size is known (zero-copy friendly for *os.File).
func ParseReaderAt(ra io.ReaderAt, size int64, ft model.Format) (*model.Document, error) {
	switch ft {
	case model.FormatDocx:
		return docx.ParseReaderAt(ra, size)
	case model.FormatXlsx:
		r := io.NewSectionReader(ra, 0, size)
		return xlsx.ParseReader(r)
	case model.FormatPptx:
		return pptx.ParseReaderAt(ra, size)
	default:
		return nil, fmt.Errorf("parse: unsupported format %q", ft)
	}
}

// ParseFileWithFormat parses when the format is already known.
func ParseFileWithFormat(path string, ft model.Format) (*model.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return ParseReaderAt(f, st.Size(), ft)
}
