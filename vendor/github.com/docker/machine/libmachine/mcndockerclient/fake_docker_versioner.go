package mcndockerclient

type FakeDockerVersioner struct {
	Version    string
	APIVersion string
	Err        error
}

func (dv *FakeDockerVersioner) DockerVersion(host DockerHost) (string, error) {
	if dv.Err != nil {
		return "", dv.Err
	}

	return dv.Version, nil
}

func (dv *FakeDockerVersioner) DockerAPIVersion(host DockerHost) (string, error) {
	if dv.Err != nil {
		return "", dv.Err
	}

	return dv.APIVersion, nil
}
