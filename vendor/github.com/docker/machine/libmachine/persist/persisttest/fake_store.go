package persisttest

import "github.com/docker/machine/libmachine/host"

type FakeStore struct {
	Hosts                                           []*host.Host
	ExistsErr, ListErr, LoadErr, RemoveErr, SaveErr error
}

func (fs *FakeStore) Exists(name string) (bool, error) {
	if fs.ExistsErr != nil {
		return false, fs.ExistsErr
	}
	for _, h := range fs.Hosts {
		if h.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (fs *FakeStore) List() ([]string, error) {
	names := []string{}
	for _, h := range fs.Hosts {
		names = append(names, h.Name)
	}
	return names, fs.ListErr
}

func (fs *FakeStore) Load(name string) (*host.Host, error) {
	if fs.LoadErr != nil {
		return nil, fs.LoadErr
	}
	for _, h := range fs.Hosts {
		if h.Name == name {
			return h, nil
		}
	}

	return nil, nil
}

func (fs *FakeStore) Remove(name string) error {
	return fs.RemoveErr
}

func (fs *FakeStore) Save(host *host.Host) error {
	return fs.SaveErr
}
