package proj2aci

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"

	"golang.org/x/tools/go/vcs"
)

type CmakeConfiguration struct {
	CommonConfiguration
	BinDir      string
	ReuseSrcDir string
	CmakeParams []string
}

type CmakePaths struct {
	CommonPaths
	src     string
	build   string
	install string
	binDir  string
}

type CmakeCustomizations struct {
	Configuration CmakeConfiguration

	paths       CmakePaths
	fullBinPath string
}

func (custom *CmakeCustomizations) Name() string {
	return "cmake"
}

func (custom *CmakeCustomizations) GetCommonConfiguration() *CommonConfiguration {
	return &custom.Configuration.CommonConfiguration
}

func (custom *CmakeCustomizations) ValidateConfiguration() error {
	if !DirExists(custom.Configuration.ReuseSrcDir) {
		return fmt.Errorf("Invalid src dir to reuse")
	}
	return nil
}

func (custom *CmakeCustomizations) GetCommonPaths() *CommonPaths {
	return &custom.paths.CommonPaths
}

func (custom *CmakeCustomizations) SetupPaths() error {
	setupReusableDir(&custom.paths.src, custom.Configuration.ReuseSrcDir, filepath.Join(custom.paths.TmpDir, "src"))
	custom.paths.build = filepath.Join(custom.paths.TmpDir, "build")
	custom.paths.install = filepath.Join(custom.paths.TmpDir, "install")
	return nil
}

func setupReusableDir(path *string, reusePath, stockPath string) {
	if path == nil {
		panic("path in setupReusableDir cannot be nil")
	}
	if reusePath != "" {
		*path = reusePath
	} else {
		*path = stockPath
	}
}

func (custom *CmakeCustomizations) GetDirectoriesToMake() []string {
	dirs := []string{
		custom.paths.build,
		custom.paths.install,
	}
	// not creating custom.paths.src, because go.vcs requires the
	// src directory to be nonexistent
	return dirs
}

func (custom *CmakeCustomizations) PrepareProject() error {
	if custom.Configuration.ReuseSrcDir == "" {
		if err := custom.createRepo(); err != nil {
			return err
		}
	}

	Info("Running cmake")
	if err := custom.runCmake(); err != nil {
		return err
	}

	Info("Running make")
	if err := custom.runMake(); err != nil {
		return err
	}

	Info("Running make install")
	if err := custom.runMakeInstall(); err != nil {
		return err
	}

	return nil
}

func (custom *CmakeCustomizations) createRepo() error {
	Info(fmt.Sprintf("Downloading %s", custom.Configuration.Project))
	repo, err := vcs.RepoRootForImportPath(custom.Configuration.Project, false)
	if err != nil {
		return err
	}
	return repo.VCS.Create(custom.paths.src, repo.Repo)
}

func (custom *CmakeCustomizations) runCmake() error {
	args := []string{"cmake"}
	args = append(args, custom.Configuration.CmakeParams...)
	args = append(args, custom.paths.src)
	return RunCmd(args, nil, custom.paths.build)
}

func (custom *CmakeCustomizations) runMake() error {
	args := []string{
		"make",
		fmt.Sprintf("-j%d", runtime.NumCPU()),
	}
	return RunCmd(args, nil, custom.paths.build)
}

func (custom *CmakeCustomizations) runMakeInstall() error {
	args := []string{
		"make",
		"install",
	}
	env := append(os.Environ(), "DESTDIR="+custom.paths.install)
	return RunCmd(args, env, custom.paths.build)
}

func (custom *CmakeCustomizations) GetPlaceholderMapping() map[string]string {
	return map[string]string{
		"<SRCPATH>":     custom.paths.src,
		"<BUILDPATH>":   custom.paths.build,
		"<INSTALLPATH>": custom.paths.install,
	}
}

func (custom *CmakeCustomizations) GetAssets(aciBinDir string) ([]string, error) {
	binaryName, err := custom.GetBinaryName()
	if err != nil {
		return nil, err
	}
	rootBinary := filepath.Join(aciBinDir, binaryName)
	return []string{GetAssetString(rootBinary, custom.fullBinPath)}, nil
}

func (custom *CmakeCustomizations) getBinDir() (string, error) {
	if custom.Configuration.BinDir != "" {
		return filepath.Join(custom.paths.install, custom.Configuration.BinDir), nil
	}
	dirs := []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}
	for _, dir := range dirs {
		path := filepath.Join(custom.paths.install, dir)
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		return path, nil
	}
	return "", fmt.Errorf("Could not find any bin directory")
}

func (custom *CmakeCustomizations) GetImageName() (*types.ACIdentifier, error) {
	imageName := custom.Configuration.Project
	if filepath.Base(imageName) == "..." {
		imageName = filepath.Dir(imageName)
		if custom.Configuration.UseBinary != "" {
			imageName += "-" + custom.Configuration.UseBinary
		}
	}
	return types.NewACIdentifier(strings.ToLower(imageName))
}

func (custom *CmakeCustomizations) GetBinaryName() (string, error) {
	if err := custom.findFullBinPath(); err != nil {
		return "", err
	}

	return filepath.Base(custom.fullBinPath), nil
}

func (custom *CmakeCustomizations) findFullBinPath() error {
	if custom.fullBinPath != "" {
		return nil
	}
	binDir, err := custom.getBinDir()
	if err != nil {
		return err
	}
	binary, err := GetBinaryName(binDir, custom.Configuration.UseBinary)
	if err != nil {
		return err
	}
	custom.fullBinPath = filepath.Join(binDir, binary)
	return nil
}

func (custom *CmakeCustomizations) GetRepoPath() (string, error) {
	return custom.paths.src, nil
}

func (custom *CmakeCustomizations) GetImageFileName() (string, error) {
	base := filepath.Base(custom.Configuration.Project)
	if base == "..." {
		base = filepath.Base(filepath.Dir(custom.Configuration.Project))
		if custom.Configuration.UseBinary != "" {
			base += "-" + custom.Configuration.UseBinary
		}
	}
	return base + schema.ACIExtension, nil
}
