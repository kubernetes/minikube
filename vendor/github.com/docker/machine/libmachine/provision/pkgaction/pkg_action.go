package pkgaction

type PackageAction int

const (
	Install PackageAction = iota
	Remove
	Upgrade
)

var packageActions = []string{
	"install",
	"remove",
	"upgrade",
}

func (s PackageAction) String() string {
	if int(s) >= 0 && int(s) < len(packageActions) {
		return packageActions[s]
	}

	return ""
}
