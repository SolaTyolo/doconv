# doconv

面向 **Office / PDF / JSON / CSV** 的 Go 库与 HTTP/CLI 服务，将文档抽取为 **Markdown**、**纯文本** 或 **JSON**。

模块：[`github.com/SolaTyolo/doconv`](https://github.com/SolaTyolo/doconv)

> **废弃说明：** 本项目已**废弃**，不再维护。如需将 Office、PDF 等文件转为 Markdown，请改用 **[MarkItDown](https://github.com/microsoft/markitdown)**。

## 功能概览

- **Office**：`.docx`、`.xlsx`、`.pptx`
- **PDF**：文本提取
- **数据**：`.json`、`.csv` / `.tsv`
- **输出**：Markdown、纯文本、结构化 JSON
- **模型**：统一的 section / 段落 / 表格，便于后续检索或 LLM 使用
- **CLI**：本地转换工具 `doconv`
- **HTTP**：`server` 提供 `POST /convert`（multipart 上传）

## 环境要求

- Go 1.22 及以上

## 代码结构

| 路径 | 说明 |
|------|------|
| `pkg/doconv` | 对外 API：`ParseFile`、`ParseReader`、`ToMarkdown`、`ToPlainText`、`ToJSON`、格式探测 |
| `internal/docx`、`xlsx`、`pptx` | 各格式解析器，输出统一模型 |
| `internal/render` | Markdown / 文本 / JSON 渲染 |
| `internal/parse` | 按扩展名或 ZIP 结构分发解析 |
| `internal/ooxml` | 共用 OOXML 逻辑（如 `docProps/core.xml`） |
| `cmd/doconv` | 命令行 |
| `cmd/server` | HTTP 服务 |

## 库用法示例

```go
doc, err := doconv.ParseFile("report.docx")
if err != nil { panic(err) }
md := doconv.ToMarkdown(doc, doconv.DefaultMarkdownOptions())
```

## 命令行

```bash
go run ./cmd/doconv -f markdown -o out.md document.docx
go run ./cmd/doconv -f json document.xlsx
```

## HTTP 服务

```bash
go run ./cmd/server -addr :8080
```

- `GET /health`：健康检查  
- `POST /convert`：表单字段 `file`（文件）、`format`（`markdown` / `text` / `json`）

## 测试

```bash
go test ./...
```

## 实现说明

- **XLSX** 使用 [excelize](https://github.com/qax-os/excelize) 读取工作表与单元格。
- **DOCX / PPTX** 使用标准库 `archive/zip` 与 `encoding/xml` 解析 OPC 包。
- **源码注释与标识符统一为英文**，便于国际化协作与工具链处理。

## 相关链接

- [MarkItDown](https://github.com/microsoft/markitdown) — **推荐替代方案**；将各类文件与 Office 文档转为 Markdown 的 Python 工具。
- [undoc](https://github.com/iyulab/undoc) — 功能更丰富的 Rust 实现（含 FFI、基准测试等）。

英文说明见 [README.md](./README.md)。
