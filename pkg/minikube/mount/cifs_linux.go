package mount

func (m *Cifs) Share() error {
	return ErrNotImplemented
}

func (m *Cifs) Unshare() error {
	return ErrNotImplemented
}

func (m *Cifs) Mount(r mountRunner) error {
	return ErrNotImplemented
}

func (m *Cifs) Unmount(r mountRunner) error {
	return ErrNotImplemented
}