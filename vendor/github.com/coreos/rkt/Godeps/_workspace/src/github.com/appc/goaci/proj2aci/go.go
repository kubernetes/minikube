package proj2aci

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

type GoConfiguration struct {
	CommonConfiguration
	GoBinary string
	GoPath   string
}

type GoPaths struct {
	CommonPaths
	project string
	realGo  string
	fakeGo  string
	goRoot  string
	goBin   string
}

type GoCustomizations struct {
	Configuration GoConfiguration

	paths GoPaths
	app   string
}

func (custom *GoCustomizations) Name() string {
	return "go"
}

func (custom *GoCustomizations) GetCommonConfiguration() *CommonConfiguration {
	return &custom.Configuration.CommonConfiguration
}

func (custom *GoCustomizations) GetCommonPaths() *CommonPaths {
	return &custom.paths.CommonPaths
}

func (custom *GoCustomizations) ValidateConfiguration() error {
	if custom.Configuration.GoBinary == "" {
		return fmt.Errorf("Go binary not found")
	}
	return nil
}

func (custom *GoCustomizations) SetupPaths() error {
	custom.paths.realGo, custom.paths.fakeGo = custom.getGoPath()

	if os.Getenv("GOPATH") != "" {
		Warn("GOPATH env var is ignored, use --go-path=\"$GOPATH\" option instead")
	}
	custom.paths.goRoot = os.Getenv("GOROOT")
	if custom.paths.goRoot != "" {
		Warn("Overriding GOROOT env var to ", custom.paths.goRoot)
	}

	projectName := getProjectName(custom.Configuration.Project)
	// Project name is path-like string with slashes, but slash is
	// not a file separator on every OS.
	custom.paths.project = filepath.Join(custom.paths.realGo, "src", filepath.Join(strings.Split(projectName, "/")...))
	custom.paths.goBin = filepath.Join(custom.paths.fakeGo, "bin")
	return nil
}

// getGoPath returns go path and fake go path. The former is either in
// /tmp (which is a default) or some other path as specified by
// --go-path parameter. The latter is always in /tmp.
func (custom *GoCustomizations) getGoPath() (string, string) {
	fakeGoPath := filepath.Join(custom.paths.TmpDir, "gopath")
	if custom.Configuration.GoPath == "" {
		return fakeGoPath, fakeGoPath
	}
	return custom.Configuration.GoPath, fakeGoPath
}

func getProjectName(project string) string {
	if filepath.Base(project) != "..." {
		return project
	}
	return filepath.Dir(project)
}

func (custom *GoCustomizations) GetDirectoriesToMake() []string {
	return []string{
		custom.paths.fakeGo,
		custom.paths.goBin,
	}
}

func (custom *GoCustomizations) PrepareProject() error {
	Info("Running go get")
	// Construct args for a go get that does a static build
	args := []string{
		"go",
		"get",
		"-a",
		custom.Configuration.Project,
	}

	env := []string{
		"GOPATH=" + custom.paths.realGo,
		"GOBIN=" + custom.paths.goBin,
		"PATH=" + os.Getenv("PATH"),
	}
	if custom.paths.goRoot != "" {
		env = append(env, "GOROOT="+custom.paths.goRoot)
	}

	cmd := exec.Cmd{
		Env:    env,
		Path:   custom.Configuration.GoBinary,
		Args:   args,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
	Debug("env: ", cmd.Env)
	Debug("running command: ", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (custom *GoCustomizations) GetPlaceholderMapping() map[string]string {
	return map[string]string{
		"<PROJPATH>": custom.paths.project,
		"<GOPATH>":   custom.paths.realGo,
	}
}

func (custom *GoCustomizations) GetAssets(aciBinDir string) ([]string, error) {
	name, err := custom.GetBinaryName()
	if err != nil {
		return nil, err
	}
	aciAsset := filepath.Join(aciBinDir, name)
	localAsset := filepath.Join(custom.paths.goBin, name)

	return []string{GetAssetString(aciAsset, localAsset)}, nil
}

func (custom *GoCustomizations) GetImageName() (*types.ACIdentifier, error) {
	imageName := custom.Configuration.Project
	if filepath.Base(imageName) == "..." {
		imageName = filepath.Dir(imageName)
		if custom.Configuration.UseBinary != "" {
			imageName += "-" + custom.Configuration.UseBinary
		}
	}
	return types.NewACIdentifier(strings.ToLower(imageName))
}

func (custom *GoCustomizations) GetBinaryName() (string, error) {
	if err := custom.findBinaryName(); err != nil {
		return "", err
	}

	return custom.app, nil
}

func (custom *GoCustomizations) findBinaryName() error {
	if custom.app != "" {
		return nil
	}
	binaryName, err := GetBinaryName(custom.paths.goBin, custom.Configuration.UseBinary)
	if err != nil {
		return err
	}
	custom.app = binaryName
	return nil
}

func (custom *GoCustomizations) GetRepoPath() (string, error) {
	return custom.paths.project, nil
}

func (custom *GoCustomizations) GetImageFileName() (string, error) {
	base := filepath.Base(custom.Configuration.Project)
	if base == "..." {
		base = filepath.Base(filepath.Dir(custom.Configuration.Project))
		if custom.Configuration.UseBinary != "" {
			base += "-" + custom.Configuration.UseBinary
		}
	}
	return base + schema.ACIExtension, nil
}
