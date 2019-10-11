package localpath

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

func chownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

// ChownToSudoUser chowns a directory to be owned by SUDO_USER
func ChownToSudoUser(dir string) error {
	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		return nil
	}

	if os.Getenv("SUDO_USER") == "" {
		return nil
	}

	username := os.Getenv("SUDO_USER")
	usr, err := user.Lookup(username)
	if err != nil {
		return errors.Wrap(err, "Error looking up user")
	}
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return errors.Wrapf(err, "Error parsing uid for user: %s", username)
	}
	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		return errors.Wrapf(err, "Error parsing gid for user: %s", username)
	}

	if err := chownR(dir, uid, gid); err != nil {
		return errors.Wrapf(err, "Error changing ownership for: %s", dir)
	}
	return nil
}

// ChownLocalDataToSudoUser chowns all local data to the SUDO_USER
func ChownLocalDataToSudoUser() error {
	for _, d := range []string{cacheDir(), configDir(), dataDir()} {
		err := ChownToSudoUser(d)
		if err != nil {
			return err
		}
	}
	return nil
}
