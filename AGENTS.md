# doconv — Cursor / Agent 开发说明

Go 库与 HTTP/CLI 服务：将 **Office**（DOCX / XLSX / PPTX）、**PDF**、**JSON**、**CSV** 抽取为统一的 section / 段落 / 表格模型，并渲染为 **Markdown**、**纯文本** 或 **JSON**。灵感来自 [undoc](https://github.com/iyulab/undoc)。

模块：[`github.com/SolaTyolo/doconv`](https://github.com/SolaTyolo/doconv)

## 关联仓库

| 仓库 | 关系 |
|------|------|
| [mcphub](../mcphub) | 通过 `internal/docparse` 调用 `pkg/doconv`，将附件转为 Markdown 供 LLM 使用 |

## 本地启动

```bash
go test ./...

# CLI
go run ./cmd/doconv -f markdown document.docx
go run ./cmd/doconv -f json -compact report.xlsx

# HTTP
go run ./cmd/server -addr :8080
curl -s -F "file=@sample.docx" -F "format=markdown" http://127.0.0.1:8080/convert
```

## 目录

| 路径 | 说明 |
|------|------|
| `pkg/doconv` | 对外 API：`ParseFile`、`ParseReader`、`ToMarkdown`、`ToPlainText`、`ToJSON`、格式探测 |
| `internal/model` | 统一文档模型（Document / Section / Element / Paragraph / Table） |
| `internal/detect` | 扩展名与 magic bytes 格式探测 |
| `internal/parse` | 按格式分发到各 parser |
| `internal/docx`、`xlsx`、`pptx`、`pdf`、`plain` | 格式解析器 |
| `internal/render` | Markdown / 文本 / JSON 渲染 |
| `internal/ooxml` | 共用 OOXML 辅助（如 `docProps/core.xml`） |
| `internal/stats` | 词数 / 字符数统计 |
| `cmd/doconv` | 命令行工具 |
| `cmd/server` | HTTP 服务（`GET /health`、`POST /convert`） |

## 架构要点

- **统一模型**：各格式 parser 输出 `model.Document`；render 层与格式无关
- **XLSX**：使用 [excelize](https://github.com/qax-os/excelize)
- **DOCX / PPTX**：标准库 `archive/zip` + `encoding/xml`，无外部 XML DOM
- **PDF**：文本页提取（非 OCR）
- **JSON / CSV**：`internal/plain` 转为单 section 文本文档
- **对外 API 稳定**：`pkg/doconv` re-export `model` 类型与常量

## 勿混淆

- 无数据库、无 gRPC / protobuf
- **源码注释与标识符统一英文**；README 可中英双语
- 新增格式：在 `internal/detect` 注册 → 新建 `internal/<format>/` parser → `internal/parse` 分发 → 更新 `SupportedExts`
- 改 public API 时同步更新 `pkg/doconv` re-export 与测试
- 未要求不 git commit；小 diff，匹配现有包布局

详见 [.cursor/DEVELOPMENT.md](.cursor/DEVELOPMENT.md)
