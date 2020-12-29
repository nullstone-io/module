package scan

import (
	"io"
	"os"
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

type ScannerItemFn func(info os.FileInfo, r io.Reader) error
