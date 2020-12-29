package scan

import (
	"io"
)

var (
	ArchiveScanners = map[string]ArchiveScanner{
		"tar.gz": TargzScanner{},
		"tgz":    TargzScanner{},
	}
)

type ArchiveScanner interface {
	Scan(r io.Reader, fn ScannerItemFn) error
}

type ScannerItemFn func(fullname string, r io.Reader) error
