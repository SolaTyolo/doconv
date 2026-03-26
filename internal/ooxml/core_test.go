package ooxml

import (
	"testing"

	"github.com/postship/doconv/internal/model"
)

func TestApplyCoreXML(t *testing.T) {
	raw := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>DocTitle</dc:title>
  <dc:creator>AuthorName</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">2020-01-15T10:30:00Z</dcterms:created>
</cp:coreProperties>`
	var m model.Metadata
	ApplyCoreXML([]byte(raw), &m)
	if m.Title != "DocTitle" || m.Author != "AuthorName" {
		t.Fatalf("metadata: %+v", m)
	}
}
