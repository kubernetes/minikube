package pkgaction

type PackageAction int

const (
	Install PackageAction = iota
	Remove
	Upgrade
	Purge
)

var packageActions = []string{
	"install",
	"remove",
	"upgrade",
	"purge",
}

func (s PackageAction) String() string {
	if int(s) >= 0 && int(s) < len(packageActions) {
		return packageActions[s]
	}

	return ""
}
