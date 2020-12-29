package scan

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

var _ ArchiveScanner = ZipScanner{}

type ZipScanner struct{}

func (s ZipScanner) Scan(r io.Reader, fn ScannerItemFn) error {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading archive: %w", err)
	}
	br := bytes.NewReader(raw)
	zr, err := zip.NewReader(br, int64(br.Len()))
	for _, fi := range zr.File {
		fr, err := fi.Open()
		if err != nil {
			return fmt.Errorf("error opening archive file %q: %w", fi.Name, err)
		}
		err = fn(fi.Name, fr)
		fr.Close()
		if err != nil {
			return fmt.Errorf("error scanning archive file %q: %w", fi.Name, err)
		}
	}
	return nil
}
