package hyperkit

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"strings"

	"github.com/hooklift/iso9660"
)

func ExtractFile(isoPath, srcPath, destPath string) error {
	iso, err := os.Open(isoPath)
	defer iso.Close()
	if err != nil {
		return err
	}

	r, err := iso9660.NewReader(iso)
	if err != nil {
		return err
	}

	f, err := findFile(r, srcPath)
	if err != nil {
		return err
	}

	dst, err := os.Create(destPath)
	defer dst.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, f.Sys().(io.Reader))
	return err
}

func ReadFile(isoPath, srcPath string) (string, error) {
	iso, err := os.Open(isoPath)
	defer iso.Close()
	if err != nil {
		return "", err
	}

	r, err := iso9660.NewReader(iso)
	if err != nil {
		return "", err
	}

	f, err := findFile(r, srcPath)
	if err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(f.Sys().(io.Reader))
	return string(contents), err
}

func findFile(r *iso9660.Reader, path string) (os.FileInfo, error) {
	for f, err := r.Next(); err != io.EOF; f, err = r.Next() {
		// Some files get an extra ',' at the end.
		if strings.TrimSuffix(f.Name(), ".") == path {
			return f, nil
		}
	}
	return nil, fmt.Errorf("Unable to find file %s.", path)
}
