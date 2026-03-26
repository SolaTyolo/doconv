# doconv

Go library and HTTP/CLI service for extracting content from Microsoft Office Open XML files (**DOCX**, **XLSX**, **PPTX**) and rendering **Markdown**, **plain text**, or **JSON**. The design goals align with the [undoc](https://github.com/iyulab/undoc) Rust project: a small, fast pipeline from OOXML to text-friendly formats.

## Features

- **Formats**: Word (`.docx`), Excel (`.xlsx`), PowerPoint (`.pptx`)
- **Outputs**: Markdown (with optional YAML frontmatter), plain text, structured JSON
- **Model**: Unified sections/paragraphs/tables suitable for downstream tools and LLM pipelines
- **CLI**: `doconv` for local conversion
- **HTTP**: `server` with `POST /convert` (multipart upload)

## Requirements

- Go 1.22+

## Project layout

| Path | Role |
|------|------|
| `pkg/doconv` | Public API: `ParseFile`, `ParseReader`, `ToMarkdown`, `ToPlainText`, `ToJSON`, format detection |
| `internal/docx`, `internal/xlsx`, `internal/pptx` | Format parsers → unified model |
| `internal/render` | Markdown / text / JSON renderers |
| `internal/parse` | Dispatch by extension or ZIP sniffing |
| `internal/ooxml` | Shared OOXML helpers (e.g. `docProps/core.xml`) |
| `cmd/doconv` | Command-line tool |
| `cmd/server` | HTTP service |

## Library usage

```go
package main

import (
	"fmt"
	"github.com/postship/doconv/pkg/doconv"
)

func main() {
	doc, err := doconv.ParseFile("report.docx")
	if err != nil {
		panic(err)
	}
	md := doconv.ToMarkdown(doc, doconv.DefaultMarkdownOptions())
	fmt.Print(md)

	jsonBytes, err := doconv.ToJSON(doc, doconv.JSONPretty)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonBytes)
}
```

## CLI

```bash
go run ./cmd/doconv -f markdown -o out.md document.docx
go run ./cmd/doconv -f json -compact document.xlsx
go run ./cmd/doconv -f text slides.pptx
```

Flags:

- `-f` — `markdown` (default), `text`, or `json`
- `-o` — output file (default: stdout)
- `-frontmatter` — YAML frontmatter for Markdown
- `-compact` — compact JSON

## HTTP server

```bash
go run ./cmd/server -addr :8080
```

- `GET /health` — JSON health check  
- `POST /convert` — form fields: `file` (multipart), `format` = `markdown` | `text` | `json`

Example:

```bash
curl -s -F "file=@sample.docx" -F "format=markdown" http://127.0.0.1:8080/convert
```

## Testing

```bash
go test ./...
```

## Implementation notes

- **XLSX** uses [excelize](https://github.com/qax-os/excelize) for reliable sheet/cell access.
- **DOCX** / **PPTX** use the standard library `archive/zip` and `encoding/xml` on the OPC packages (no external XML DOM).
- Comments and identifiers in source code are **English** for consistency across tooling.

## License

See your organization’s policy; the scaffold is suitable for MIT or BSD-style licensing if you add a `LICENSE` file.

## See also

- [undoc](https://github.com/iyulab/undoc) — Rust implementation with additional features (FFI, benchmarks, richer cleanup presets).

Chinese documentation: [README.zh-CN.md](./README.zh-CN.md).
