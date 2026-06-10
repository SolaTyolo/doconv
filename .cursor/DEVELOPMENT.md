# 本地开发（Cursor）

## 环境要求

- Go 1.22+

无 `.env` 或外部服务依赖；CLI 与 HTTP server 均可直接 `go run`。

## 启动与调试

```bash
go test ./...
go test ./internal/docx/... -v -run TestName   # 单包 / 单测

# CLI
go run ./cmd/doconv -f markdown -o out.md document.docx
go run ./cmd/doconv -f text slides.pptx
go run ./cmd/doconv -f json -compact data.json

# HTTP server
go run ./cmd/server -addr :8080
curl http://127.0.0.1:8080/health
curl -s -F "file=@sample.docx" -F "format=markdown" http://127.0.0.1:8080/convert
```

## 解析流程

```
输入（path / io.Reader）
  → internal/detect（扩展名或 magic bytes）
  → internal/parse（分发）
  → internal/{docx,xlsx,pptx,pdf,plain}/（格式 parser）
  → model.Document
  → internal/render（ToMarkdown / ToPlainText / ToJSON）
```

- `ParseReader(r, filenameHint)`：优先用 `filenameHint` 扩展名；否则读全量后 `FromBytes` 嗅探
- OOXML（docx/xlsx/pptx）：ZIP 结构 + 各包内 XML
- `ParseReaderAt`：已知格式 + 随机访问 reader 时使用

## 统一模型

| 类型 | 用途 |
|------|------|
| `Document` | 根；含 Format、Metadata、Sections、Stats |
| `Section` | Word 节 / Excel 工作表 / PPT 幻灯片 |
| `Element` | `paragraph` 或 `table` |
| `Paragraph` | `[]Run` + 可选 `OutlineLevel`（标题层级提示） |
| `Table` | `[][]TableCell`，每 cell 可含多段落 |

渲染选项见 `internal/render`：`MarkdownOptions`（frontmatter、table mode）、`TextOptions`、`JSONFormat`。

## 支持格式

| 扩展名 | Format | Parser |
|--------|--------|--------|
| `.docx` | docx | `internal/docx` |
| `.xlsx` | xlsx | `internal/xlsx` |
| `.pptx` | pptx | `internal/pptx` |
| `.pdf` | pdf | `internal/pdf` |
| `.json` | json | `internal/plain` |
| `.csv` / `.tsv` | csv | `internal/plain` |

## 新增格式 checklist

1. `internal/model`：添加 `FormatXxx` 常量（若需要）
2. `internal/detect`：`FromPath` / `FromBytes` 识别
3. `internal/<format>/`：实现 parser，返回 `*model.Document`
4. `internal/parse`：在 `ParseFile` / `ParseReader` / `parseBytesWithFormat` 中分发
5. `pkg/doconv`：re-export 新 Format 常量（若 public）
6. `internal/*_test.go`：至少一个 round-trip 或 golden 测试
7. 更新 `README.md` / `README.zh-CN.md`

## mcphub 集成

mcphub 通过 `replace` 或 module path 引用本库，封装于 `internal/docparse`：

```go
doc, err := doconv.ParseReader(bytes.NewReader(data), filename)
md := doconv.ToMarkdown(doc, doconv.DefaultMarkdownOptions())
```

改 `pkg/doconv` 签名或行为时需同步检查 mcphub 的 `internal/docparse`。

## 改代码时的模式

1. **Public API**：仅 `pkg/doconv` 对外暴露；内部逻辑放 `internal/`
2. **Parser**：各格式独立包，输出统一 `model.Document`，不直接写 Markdown
3. **Render**：格式无关；表格模式见 `TableMarkdown` / `TableHTML` / `TableASCII`
4. **测试**：`*_test.go` 与源文件同包；优先 table-driven 与 fixture 文件

保持小 diff，与现有 `internal/<format>/` 拆分风格一致。

## 常见坑

- `ParseReader` 会将整个 stream 读入内存；大文件场景用 `ParseReaderAt` 或 CLI 直读路径
- HTTP `/convert` 限制 multipart 32 MiB（与 mcphub `docparse.MaxDocumentBytes` 对齐时注意）
- CSV 解析用 `LazyQuotes`；TSV 通过 `.tsv` 扩展名切换 tab 分隔
- DOCX 标题依赖 `OutlineLevel`，非所有文档都有规范 outline
- 无 Makefile；直接用 `go test` / `go run`
