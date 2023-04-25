package serviceaction

type ServiceAction int

const (
	Restart ServiceAction = iota
	Start
	Stop
	Enable
	Disable
	DaemonReload
)

var serviceActions = []string{
	"restart",
	"start",
	"stop",
	"enable",
	"disable",
	"daemon-reload",
}

func (s ServiceAction) String() string {
	if int(s) >= 0 && int(s) < len(serviceActions) {
		return serviceActions[s]
	}

	return ""
}
