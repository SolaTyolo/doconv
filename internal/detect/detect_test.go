package detect

import "testing"

func TestFromPathExtended(t *testing.T) {
	cases := map[string]string{
		"report.pdf":      "pdf",
		"data.json":       "json",
		"sheet.csv":       "csv",
		"tab.tsv":  "csv",
		"doc.docx": "docx",
	}
	for path, want := range cases {
		got, err := FromPath(path)
		if err != nil {
			t.Fatalf("FromPath(%q): %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("FromPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestFromBytesPDF(t *testing.T) {
	got, err := FromBytes([]byte("%PDF-1.4\n"))
	if err != nil || got != "pdf" {
		t.Fatalf("FromBytes pdf: got=%q err=%v", got, err)
	}
}

func TestFromBytesJSON(t *testing.T) {
	got, err := FromBytes([]byte(`{"a":1}`))
	if err != nil || got != "json" {
		t.Fatalf("FromBytes json: got=%q err=%v", got, err)
	}
}
