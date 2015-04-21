// Package internet gives access to information about internet by
// downloading/parsing/putting ripe BGP dumps and cidr-report.org data into
// redis databases. Redis query clients are also supplied.
//
// This is a work in progress, the API's are unstable.
package internet

import (
	"os"
	"path/filepath"
)

var dataDir = filepath.Join(os.TempDir(), "internet")

// SetDataDir sets the storage directory for downloads cache and temporary files.
func SetDataDir(path string) {
	dataDir = path
}
