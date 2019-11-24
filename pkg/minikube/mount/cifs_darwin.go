package mount

func (m *Cifs) Share(config MountConfig) error {
	return MountNotImplementedError
}

func (m *Cifs) Unshare(config MountConfig) error {
	return MountNotImplementedError
}

func (m *Cifs) Mount(r mountRunner, config MountConfig) error {
	return MountNotImplementedError
}

func (m *Cifs) Unmount(r mountRunner, config MountConfig) error {
	return MountNotImplementedError
}