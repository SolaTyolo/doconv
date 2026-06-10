// Package ooxml contains shared Office Open XML helpers (metadata, zip).
package ooxml

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/SolaTyolo/doconv/internal/model"
)

// coreProps maps docProps/core.xml with explicit namespaces used by Office.
type coreProps struct {
	XMLName  xml.Name `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties coreProperties"`
	Title    string   `xml:"http://purl.org/dc/elements/1.1/ title"`
	Subject  string   `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Creator  string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Created  string   `xml:"http://purl.org/dc/terms/ created"`
	Modified string   `xml:"http://purl.org/dc/terms/ modified"`
}

// ApplyCoreXML fills metadata fields from docProps/core.xml bytes.
func ApplyCoreXML(data []byte, meta *model.Metadata) {
	var c coreProps
	if err := xml.Unmarshal(data, &c); err != nil {
		return
	}
	meta.Title = strings.TrimSpace(c.Title)
	meta.Author = strings.TrimSpace(c.Creator)
	meta.Subject = strings.TrimSpace(c.Subject)
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(c.Created)); err == nil {
		meta.Created = t
	}
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(c.Modified)); err == nil {
		meta.Modified = t
	}
}
