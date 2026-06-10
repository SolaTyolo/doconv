package doconv

import "github.com/SolaTyolo/doconv/internal/detect"

// DetectFromPath returns the format from file extension.
func DetectFromPath(path string) (Format, error) {
	ft, err := detect.FromPath(path)
	return Format(ft), err
}

// DetectFromBytes inspects a ZIP/OOXML payload.
func DetectFromBytes(data []byte) (Format, error) {
	ft, err := detect.FromBytes(data)
	return Format(ft), err
}

// ErrUnknownFormat is returned when the format cannot be determined.
var ErrUnknownFormat = detect.ErrUnknownFormat
