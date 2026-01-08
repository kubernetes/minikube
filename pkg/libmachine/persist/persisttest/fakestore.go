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
	if fs.RemoveErr != nil {
		return fs.RemoveErr
	}
	for i, h := range fs.Hosts {
		if h.Name == name {
			fs.Hosts = append(fs.Hosts[:i], fs.Hosts[i+1:]...)
			return nil
		}
	}
	return nil
}

func (fs *FakeStore) Save(host *host.Host) error {
	if fs.SaveErr == nil {
		fs.Hosts = append(fs.Hosts, host)
		return nil
	}
	return fs.SaveErr
}
