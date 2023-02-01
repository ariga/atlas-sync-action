package main

import (
	"bytes"
	"fmt"
	"io"

	"ariga.io/atlas/sql/migrate"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/tools/txtar"
)

func main() {
	githubactions.Noticef("Hello, GitHub Actions")
}

// Archive returns a txtar archive of the given migration directory.
func Archive(path string) (string, error) {
	d, err := migrate.NewLocalDir(path)
	files, err := d.Files()
	if err != nil {
		return "", err
	}
	arc := &txtar.Archive{}
	for _, f := range files {
		arc.Files = append(arc.Files, txtar.File{
			Name: f.Name(),
			Data: f.Bytes(),
		})
	}
	sumf, err := d.Open(migrate.HashFileName)
	if err != nil {
		return "", fmt.Errorf("opening sumfile: %w", err)
	}
	curS, err := io.ReadAll(sumf)
	if err != nil {
		return "", fmt.Errorf("reading sumfile: %w", err)
	}
	sum, err := d.Checksum()
	if err != nil {
		return "", err
	}
	wantS, err := sum.MarshalText()
	if err != nil {
		return "", err
	}
	if !bytes.Equal(curS, wantS) {
		return "", migrate.ErrChecksumMismatch
	}
	arc.Files = append(arc.Files, txtar.File{
		Name: migrate.HashFileName,
		Data: wantS,
	})
	return string(txtar.Format(arc)), nil
}
