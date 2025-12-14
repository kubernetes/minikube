package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/hooklift/iso9660"
)

// Version holds the CLI version and is set in compilation time.
var Version string

func main() {
	usage := `ISO9660 extractor.
Usage:
  iso9660 <image-path> <destination-path>
  iso9660 -h | --help
  iso9660 --version
`

	args, err := docopt.Parse(usage, nil, true, Version, false)
	if err != nil {
		panic(err)
	}

	file, err := os.Open(args["<image-path>"].(string))
	if err != nil {
		panic(err)
	}

	r, err := iso9660.NewReader(file)
	if err != nil {
		panic(err)
	}

	destPath := args["<destination-path>"].(string)
	if destPath == "" {
		destPath = "."
	}

	for {
		f, err := r.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		fp := filepath.Join(destPath, f.Name())
		if f.IsDir() {
			if err := os.MkdirAll(fp, f.Mode()); err != nil {
				panic(err)
			}
			continue
		}

		parentDir, _ := filepath.Split(fp)
		if err := os.MkdirAll(parentDir, f.Mode()); err != nil {
			panic(err)
		}

		fmt.Printf("Extracting %s...\n", fp)

		freader := f.Sys().(io.Reader)
		ff, err := os.Create(fp)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := ff.Close(); err != nil {
				panic(err)
			}
		}()

		if err := ff.Chmod(f.Mode()); err != nil {
			panic(err)
		}

		if _, err := io.Copy(ff, freader); err != nil {
			panic(err)
		}
	}
}
