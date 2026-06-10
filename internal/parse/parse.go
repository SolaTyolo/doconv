// Package parse dispatches to format-specific parsers from paths or byte streams.
package parse

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SolaTyolo/doconv/internal/detect"
	"github.com/SolaTyolo/doconv/internal/docx"
	"github.com/SolaTyolo/doconv/internal/model"
	"github.com/SolaTyolo/doconv/internal/pdf"
	"github.com/SolaTyolo/doconv/internal/plain"
	"github.com/SolaTyolo/doconv/internal/pptx"
	"github.com/SolaTyolo/doconv/internal/xlsx"
)

// ParseFile detects format from extension and parses the file.
func ParseFile(path string) (*model.Document, error) {
	ft, err := detect.FromPath(path)
	if err != nil {
		return nil, err
	}
	if isPlainFormat(ft) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return parsePlainBytes(data, ft, path)
	}
	switch ft {
	case model.FormatDocx:
		return docx.ParseFile(path)
	case model.FormatXlsx:
		return xlsx.ParseFile(path)
	case model.FormatPptx:
		return pptx.ParseFile(path)
	case model.FormatPDF:
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return pdf.ParseReader(f)
	default:
		return nil, fmt.Errorf("parse: unsupported format %q", ft)
	}
}

// ParseReader reads the entire stream into memory, detects type, and parses.
func ParseReader(r io.Reader, filenameHint string) (*model.Document, error) {
	if ft, err := detect.FromPath(filenameHint); err == nil {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if isPlainFormat(ft) {
			return parsePlainBytes(data, ft, filenameHint)
		}
		if ft == model.FormatPDF {
			return pdf.ParseBytes(data)
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
	if ft == model.FormatPDF {
		return pdf.ParseBytes(data)
	}
	if isPlainFormat(ft) {
		return parsePlainBytes(data, ft, filenameHint)
	}
	return parseBytesWithFormat(data, ft)
}

func isPlainFormat(ft model.Format) bool {
	switch ft {
	case model.FormatJSON, model.FormatCSV:
		return true
	default:
		return false
	}
}

func parsePlainBytes(data []byte, ft model.Format, filename string) (*model.Document, error) {
	switch ft {
	case model.FormatJSON:
		return plain.ParseJSON(data)
	case model.FormatCSV:
		return plain.ParseCSV(data, filename)
	default:
		return nil, fmt.Errorf("parse: unsupported plain format %q", ft)
	}
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

// ParseReaderAt parses from a random-access reader when size is known.
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
	if isPlainFormat(ft) || ft == model.FormatPDF {
		return ParseFile(path)
	}
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

// SupportedFilename reports whether filename/path is supported.
func SupportedFilename(name string) bool {
	return detect.SupportedPath(name)
}

// SupportedExts returns a human-readable list of supported inputs.
func SupportedExts() string {
	return strings.Join([]string{
		".docx", ".xlsx", ".pptx",
		".pdf", ".json", ".csv", ".tsv",
	}, ", ")
}
