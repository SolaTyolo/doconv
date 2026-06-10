// Command doconv converts Office documents to Markdown, plain text, or JSON.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/SolaTyolo/doconv/pkg/doconv"
)

func main() {
	outPath := flag.String("o", "", "output file (default: stdout)")
	format := flag.String("f", "markdown", "output format: markdown, text, json")
	jsonCompact := flag.Bool("compact", false, "compact JSON (only for -f json)")
	frontmatter := flag.Bool("frontmatter", false, "YAML frontmatter in markdown")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: doconv [-f markdown|text|json] [-o out] <file.docx|xlsx|pptx>")
		os.Exit(2)
	}
	path := args[0]
	doc, err := doconv.ParseFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var data []byte
	switch strings.ToLower(*format) {
	case "markdown", "md":
		opt := doconv.DefaultMarkdownOptions()
		opt.Frontmatter = *frontmatter
		data = []byte(doconv.ToMarkdown(doc, opt))
	case "text", "txt":
		data = []byte(doconv.ToPlainText(doc, doconv.DefaultTextOptions()))
	case "json":
		jf := doconv.JSONPretty
		if *jsonCompact {
			jf = doconv.JSONCompact
		}
		data, err = doconv.ToJSON(doc, jf)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown format:", *format)
		os.Exit(2)
	}
	if *outPath != "" {
		if err := os.WriteFile(*outPath, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	os.Stdout.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		os.Stdout.WriteString("\n")
	}
}
