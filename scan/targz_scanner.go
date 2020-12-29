package scan

import (
	"compress/gzip"
	"fmt"
	"io"
)

var _ ArchiveScanner = TargzScanner{}

type TargzScanner struct{}

func (s TargzScanner) Scan(r io.Reader, fn ScannerItemFn) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gr.Close()

	return (TarScanner{}).Scan(gr, fn)
}
