package pdf

import "testing"

func TestExtractTextSimple(t *testing.T) {
	data := []byte(`%PDF-1.4
1 0 obj
<<>>
endobj
stream
BT (Hello PDF) Tj ET
endstream
`)
	text, err := ExtractText(data)
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello PDF" {
		t.Fatalf("got %q", text)
	}
}
