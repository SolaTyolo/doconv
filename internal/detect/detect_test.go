package detect

import (
	"testing"

	"github.com/postship/doconv/internal/model"
)

func TestFromPath(t *testing.T) {
	ft, err := FromPath("/tmp/x.docx")
	if err != nil || ft != model.FormatDocx {
		t.Fatalf("docx: %v %q", err, ft)
	}
	_, err = FromPath("unknown.bin")
	if err != ErrUnknownFormat {
		t.Fatalf("expected ErrUnknownFormat, got %v", err)
	}
}
