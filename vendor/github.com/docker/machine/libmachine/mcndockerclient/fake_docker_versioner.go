package mcndockerclient

type FakeDockerVersioner struct {
	Version string
	Err     error
}

func (dv *FakeDockerVersioner) DockerVersion(host DockerHost) (string, error) {
	if dv.Err != nil {
		return "", dv.Err
	}

	return dv.Version, nil
}
