package iso9660

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hooklift/assert"
)

func TestNewReader(t *testing.T) {
	image, err := os.Open("./fixtures/test.iso")
	defer image.Close()
	r, err := NewReader(image)
	assert.Ok(t, err)
	// Test first half of primary volume descriptor
	assert.Equals(t, "CD001", string(r.pvd.StandardID[:]))
	assert.Equals(t, 1, int(r.pvd.Type))
	assert.Equals(t, 1, int(r.pvd.Version))
	assert.Equals(t, "Mac OS X", strings.TrimSpace(string(r.pvd.SystemID[:])))
	assert.Equals(t, "my-vol-id", strings.TrimSpace(string(r.pvd.ID[:])))
	assert.Equals(t, 181, int(r.pvd.VolumeSpaceSizeBE))
	assert.Equals(t, 1, int(r.pvd.VolumeSetSizeBE))
	assert.Equals(t, 1, int(r.pvd.VolumeSeqNumberBE))
	assert.Equals(t, 2048, int(r.pvd.LogicalBlkSizeBE))
	assert.Equals(t, 46, int(r.pvd.PathTableSizeBE))
	assert.Equals(t, 21, int(r.pvd.LocPathTableBE))
	assert.Equals(t, 0, int(r.pvd.LocOptPathTableBE))
	// Test root directory record values
	assert.Equals(t, 0, int(r.pvd.DirectoryRecord.ExtendedAttrLen))
	assert.Equals(t, 23, int(r.pvd.DirectoryRecord.ExtentLocationBE))
	assert.Equals(t, 2048, int(r.pvd.DirectoryRecord.ExtentLengthBE))
	assert.Equals(t, 2, int(r.pvd.DirectoryRecord.FileFlags))
	assert.Equals(t, 0, int(r.pvd.DirectoryRecord.FileUnitSize))
	assert.Equals(t, 0, int(r.pvd.DirectoryRecord.InterleaveGapSize))
	assert.Equals(t, 1, int(r.pvd.DirectoryRecord.VolumeSeqNumberBE))
	assert.Equals(t, 1, int(r.pvd.DirectoryRecord.FileIDLength))
	// Test second half of primary volume descriptor
	assert.Equals(t, "my-vol-id", strings.TrimSpace(string(r.pvd.ID[:])))
	assert.Equals(t, "test-volset-id", strings.TrimSpace(string(r.pvd.VolumeSetID[:])))
	assert.Equals(t, "hooklift", strings.TrimSpace(string(r.pvd.PublisherID[:])))
	assert.Equals(t, "hooklift", strings.TrimSpace(string(r.pvd.DataPreparerID[:])))
	assert.Equals(t, "MKISOFS ISO9660/HFS/UDF FILESYSTEM BUILDER & CDRECORD CD/DVD/BluRay CREATOR (C) 1993 E.YOUNGDALE (C) 1997 J.PEARSON/J.SCHILLING", strings.TrimSpace(string(r.pvd.AppID[:])))
	assert.Equals(t, 1, int(r.pvd.FileStructVersion))
}

func TestUnpacking(t *testing.T) {
	image, err := os.Open("./fixtures/test.iso")
	defer image.Close()
	reader, err := NewReader(image)
	assert.Ok(t, err)

	tests := []struct {
		name    string
		isDir   bool
		content string
	}{
		{"/dir1", true, ""},
		{"/dir2", true, ""},
		{"/file.txt", false, "hola amigo\n"},
		{"/dir1/hello.txt", false, "hello there!"},
		{"/dir2/dir3", true, ""},
		{"/dir2/dir3/blah.txt", false, "do you feel me?\n"},
	}

	count := 0
	for {
		fi, err := reader.Next()
		if err == io.EOF {
			break
		}
		assert.Ok(t, err)

		f := fi.(*File)
		//fmt.Printf("%s\n", tests[count].name)
		assert.Equals(t, tests[count].name, f.Name())
		assert.Equals(t, tests[count].isDir, f.IsDir())

		rawBytes := f.Sys()
		if !f.IsDir() {
			assert.Cond(t, rawBytes != nil, "when it is file, content should not be nil")
			content, err := ioutil.ReadAll(rawBytes.(io.Reader))
			assert.Ok(t, err)
			//fmt.Printf("%s -> %s\n", tests[count].name, content)
			assert.Equals(t, tests[count].content, string(content[:]))
		} else {
			assert.Equals(t, nil, rawBytes)
		}
		count++
	}
	assert.Equals(t, 6, count)
}

// func TestBigImage(t *testing.T) {
// 	image, err := os.Open("./fixtures/test.iso.arch")
// 	defer image.Close()
// 	reader, err := NewReader(image)
// 	assert.Ok(t, err)
//
// 	count := 0
// 	for {
// 		fi, err := reader.Next()
// 		if err == io.EOF {
// 			break
// 		}
// 		assert.Ok(t, err)
//
// 		f := fi.(*File)
// 		fmt.Printf("%s\n", f.Name())
// 		count++
// 	}
//
// 	assert.Equals(t, 121, count)
// }

func ExampleReader() {
	file, err := os.Open("archlinux.iso")
	if err != nil {
		panic(err)
	}

	r, err := NewReader(file)
	if err != nil {
		panic(err)
	}

	destPath := "tmp"
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
