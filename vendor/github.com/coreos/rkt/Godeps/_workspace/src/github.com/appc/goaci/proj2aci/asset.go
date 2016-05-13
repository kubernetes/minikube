package proj2aci

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GetAssetString returns a properly formatted asset string.
func GetAssetString(aciAsset, localAsset string) string {
	return getAssetString(aciAsset, localAsset)
}

// PrepareAssets copies given assets to ACI rootfs directory. It also
// tries to copy required shared libraries if an asset is a
// dynamically linked executable or library. placeholderMapping maps
// placeholders (like "<INSTALLDIR>") to actual paths (usually
// something inside temporary directory).
func PrepareAssets(assets []string, rootfs string, placeholderMapping map[string]string) error {
	newAssets := assets
	processedAssets := make(map[string]struct{})
	for len(newAssets) > 0 {
		assetsToProcess := newAssets
		newAssets = nil
		for _, asset := range assetsToProcess {
			splitAsset := filepath.SplitList(asset)
			if len(splitAsset) != 2 {
				return fmt.Errorf("Malformed asset option: '%v' - expected two absolute paths separated with %v", asset, listSeparator())
			}
			evalSrc, srcErr := evalPath(splitAsset[0])
			if srcErr != nil {
				return fmt.Errorf("Could not evaluate symlinks in source asset %q: %v", evalSrc, srcErr)
			}
			evalDest, destErr := evalPath(splitAsset[1])
			asset = getAssetString(evalSrc, evalDest)
			Debug("Processing asset:", asset)
			if _, ok := processedAssets[asset]; ok {
				Debug("  skipped")
				continue
			}
			additionalAssets, err := processAsset(asset, rootfs, placeholderMapping)
			if destErr != nil {
				evalDest, destErr = evalPath(evalDest)
				if destErr != nil {
					return fmt.Errorf("Could not evaluate symlinks in destination asset %q, even after it was copied to: %v", evalDest, destErr)
				}
				asset = getAssetString(evalSrc, evalDest)
			}
			if err != nil {
				return err
			}
			processedAssets[asset] = struct{}{}
			newAssets = append(newAssets, additionalAssets...)
		}
	}
	return nil
}

func evalPath(path string) (string, error) {
	symlinkDir := filepath.Dir(path)
	symlinkBase := filepath.Base(path)
	realSymlinkDir, err := filepath.EvalSymlinks(symlinkDir)
	if err != nil {
		return path, err
	}
	return filepath.Join(realSymlinkDir, symlinkBase), nil
}

// processAsset validates an asset, replaces placeholders with real
// paths and does the copying. It may return additional assets to be
// processed when asset is an executable or a library.
func processAsset(asset, rootfs string, placeholderMapping map[string]string) ([]string, error) {
	splitAsset := filepath.SplitList(asset)
	if len(splitAsset) != 2 {
		return nil, fmt.Errorf("Malformed asset option: '%v' - expected two absolute paths separated with %v", asset, listSeparator())
	}
	ACIAsset := replacePlaceholders(splitAsset[0], placeholderMapping)
	localAsset := replacePlaceholders(splitAsset[1], placeholderMapping)
	if err := validateAsset(ACIAsset, localAsset); err != nil {
		return nil, err
	}
	ACIAssetSubPath := filepath.Join(rootfs, filepath.Dir(ACIAsset))
	err := os.MkdirAll(ACIAssetSubPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed to create directory tree for asset '%v': %v", asset, err)
	}
	err = copyTree(localAsset, filepath.Join(rootfs, ACIAsset))
	if err != nil {
		return nil, fmt.Errorf("Failed to copy assets for %q: %v", asset, err)
	}
	additionalAssets, err := getSoLibs(localAsset)
	if err != nil {
		return nil, fmt.Errorf("Failed to get dependent assets for %q: %v", localAsset, err)
	}
	// HACK! if we are copying libc then try to copy libnss_* libs
	// (glibc name service switch libs used in networking)
	if matched, err := filepath.Match("libc.*", filepath.Base(localAsset)); err == nil && matched {
		toGlob := filepath.Join(filepath.Dir(localAsset), "libnss_*")
		if matches, err := filepath.Glob(toGlob); err == nil && len(matches) > 0 {
			matchesAsAssets := make([]string, 0, len(matches))
			for _, f := range matches {
				matchesAsAssets = append(matchesAsAssets, getAssetString(f, f))
			}
			additionalAssets = append(additionalAssets, matchesAsAssets...)
		}
	}
	return additionalAssets, nil
}

func replacePlaceholders(path string, placeholderMapping map[string]string) string {
	Debug("Processing path: ", path)
	newPath := path
	for placeholder, replacement := range placeholderMapping {
		newPath = strings.Replace(newPath, placeholder, replacement, -1)
	}
	Debug("Processed path: ", newPath)
	return newPath
}

func validateAsset(ACIAsset, localAsset string) error {
	if !filepath.IsAbs(ACIAsset) {
		return fmt.Errorf("Wrong ACI asset: '%v' - ACI asset has to be absolute path", ACIAsset)
	}
	if !filepath.IsAbs(localAsset) {
		return fmt.Errorf("Wrong local asset: '%v' - local asset has to be absolute path", localAsset)
	}
	fi, err := os.Lstat(localAsset)
	if err != nil {
		return fmt.Errorf("Error stating %v: %v", localAsset, err)
	}
	mode := fi.Mode()
	if mode.IsDir() || mode.IsRegular() || isSymlink(mode) {
		return nil
	}
	return fmt.Errorf("Can't handle local asset %v - not a file, not a dir, not a symlink", fi.Name())
}

func copyTree(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rootLess := path[len(src):]
		target := filepath.Join(dest, rootLess)
		mode := info.Mode()
		switch {
		case mode.IsDir():
			err := os.Mkdir(target, mode.Perm())
			if err != nil {
				return err
			}
		case mode.IsRegular():
			if err := copyRegularFile(path, target); err != nil {
				return err
			}
		case isSymlink(mode):
			if err := copySymlink(path, target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unsupported node %q in assets, only regular files, directories and symlinks are supported.", path, mode.String())
		}
		return nil
	})
}

func copyRegularFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}
	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if err := destFile.Chmod(fi.Mode().Perm()); err != nil {
		return err
	}
	return nil
}

func copySymlink(src, dest string) error {
	symTarget, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.Symlink(symTarget, dest); err != nil {
		return err
	}
	return nil
}

// getSoLibs tries to run ldd on given path and to process its output
// to get a list of shared libraries to copy. This list is returned as
// an array of assets.
//
// man ldd says that running ldd on untrusted executables is dangerous
// (it might run an executable to get the libraries), so possibly this
// should be replaced with objdump. Problem with objdump is that it
// just gives library names, while ldd gives absolute paths to those
// libraries - to use objdump we need to know the $libdir.
func getSoLibs(path string) ([]string, error) {
	assets := []string{}
	buf := new(bytes.Buffer)
	args := []string{
		"ldd",
		path,
	}
	if err := RunCmdFull("", args, nil, "", buf, nil); err != nil {
		if _, ok := err.(CmdFailedError); !ok {
			return nil, err
		}
	} else {
		re := regexp.MustCompile(`(?m)^\t(?:\S+\s+=>\s+)?(\/\S+)\s+\([0-9a-fA-Fx]+\)$`)
		for _, matches := range re.FindAllStringSubmatch(string(buf.Bytes()), -1) {
			lib := matches[1]
			if lib == "" {
				continue
			}
			symlinkedAssets, err := getSymlinkedAssets(lib)
			if err != nil {
				return nil, err
			}
			assets = append(assets, symlinkedAssets...)
		}
	}
	return assets, nil
}

// getSymlinkedAssets returns an array of many assets if given path is
// a symlink - useful for getting shared libraries, which are often
// surrounded with a bunch of symlinks.
func getSymlinkedAssets(path string) ([]string, error) {
	assets := []string{}
	maxLevels := 100
	levels := maxLevels
	for {
		if levels < 1 {
			return nil, fmt.Errorf("Too many levels of symlinks (>$d)", maxLevels)
		}
		fi, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		asset := getAssetString(path, path)
		assets = append(assets, asset)
		if !isSymlink(fi.Mode()) {
			break
		}
		symTarget, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}
		if filepath.IsAbs(symTarget) {
			path = symTarget
		} else {
			path = filepath.Join(filepath.Dir(path), symTarget)
		}
		levels--
	}
	return assets, nil
}

func getAssetString(aciAsset, localAsset string) string {
	return fmt.Sprintf("%s%s%s", aciAsset, listSeparator(), localAsset)
}

func isSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink == os.ModeSymlink
}
