// Package pdf extracts plain text from PDF files (best-effort, no CGO).
package pdf

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/SolaTyolo/doconv/internal/model"
)

var (
	pdfLiteralRe = regexp.MustCompile(`\((?:\\.|[^\\()])*?\)`)
	pdfHexRe     = regexp.MustCompile(`<[0-9A-Fa-f\s]+>`)
)

func ParseReader(r io.Reader) (*model.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseBytes(data)
}

func ParseBytes(data []byte) (*model.Document, error) {
	text, err := ExtractText(data)
	if err != nil {
		return nil, err
	}
	return model.TextDocument(model.FormatPDF, "PDF", text), nil
}

// ExtractText pulls visible text literals from PDF content streams.
func ExtractText(data []byte) (string, error) {
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		return "", fmt.Errorf("not a PDF file")
	}

	seen := make(map[string]struct{})
	var parts []string

	for _, m := range pdfLiteralRe.FindAll(data, -1) {
		raw := string(m[1 : len(m)-1])
		decoded := strings.TrimSpace(decodeLiteral(raw))
		if decoded == "" || !utf8.ValidString(decoded) {
			continue
		}
		if _, ok := seen[decoded]; ok {
			continue
		}
		seen[decoded] = struct{}{}
		parts = append(parts, decoded)
	}

	for _, m := range pdfHexRe.FindAll(data, -1) {
		raw := strings.TrimSpace(string(m[1 : len(m)-1]))
		decoded := strings.TrimSpace(decodeHex(raw))
		if decoded == "" {
			continue
		}
		if _, ok := seen[decoded]; ok {
			continue
		}
		seen[decoded] = struct{}{}
		parts = append(parts, decoded)
	}

	body := strings.TrimSpace(strings.Join(parts, "\n"))
	if body == "" {
		return "", fmt.Errorf("no extractable text in PDF")
	}
	return body, nil
}

func decodeLiteral(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case '(', ')', '\\':
			b.WriteByte(s[i])
		default:
			if i+2 < len(s) && s[i] >= '0' && s[i] <= '7' {
				var v rune
				fmt.Sscanf(s[i:i+3], "%o", &v)
				b.WriteRune(v)
				i += 2
			} else {
				b.WriteByte(s[i])
			}
		}
	}
	return b.String()
}

func decodeHex(hex string) string {
	hex = strings.ReplaceAll(hex, " ", "")
	if len(hex)%2 == 1 {
		hex += "0"
	}
	var b strings.Builder
	for i := 0; i+1 < len(hex); i += 2 {
		var v byte
		if _, err := fmt.Sscanf(hex[i:i+2], "%02x", &v); err != nil {
			continue
		}
		if v == 0 {
			continue
		}
		b.WriteByte(v)
	}
	return b.String()
}
