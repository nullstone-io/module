package scan

import (
	"archive/tar"
	"fmt"
	"io"
)

var _ ArchiveScanner = TarScanner{}

type TarScanner struct{}

func (s TarScanner) Scan(r io.Reader, fn ScannerItemFn) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error advancing to next item in tar: %w", err)
		}
		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			// don't unpack extended headers as files
			continue
		}
		if err := fn(hdr.Name, tr); err != nil {
			return fmt.Errorf("error scanning %q: %w", hdr.Name, err)
		}
	}
	return nil
}
