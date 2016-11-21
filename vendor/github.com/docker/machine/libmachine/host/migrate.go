package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/version"
)

var (
	errConfigFromFuture = errors.New("config version is from the future -- you should upgrade your Docker Machine client")
)

type RawDataDriver struct {
	*none.Driver
	Data []byte // passed directly back when invoking json.Marshal on this type
}

func (r *RawDataDriver) MarshalJSON() ([]byte, error) {
	return r.Data, nil
}

func (r *RawDataDriver) UnmarshalJSON(data []byte) error {
	r.Data = data
	return nil
}

func getMigratedHostMetadata(data []byte) (*Metadata, error) {
	// HostMetadata is for a "first pass" so we can then load the driver
	var (
		hostMetadata *MetadataV0
	)

	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return &Metadata{}, err
	}

	migratedHostMetadata := MigrateHostMetadataV0ToHostMetadataV1(hostMetadata)

	return migratedHostMetadata, nil
}

func MigrateHost(h *Host, data []byte) (*Host, bool, error) {
	var (
		migrationNeeded    = false
		migrationPerformed = false
		hostV1             *V1
		hostV2             *V2
	)

	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return nil, false, err
	}

	globalStorePath := filepath.Dir(filepath.Dir(migratedHostMetadata.HostOptions.AuthOptions.StorePath))

	driver := &RawDataDriver{none.NewDriver(h.Name, globalStorePath), nil}

	if migratedHostMetadata.ConfigVersion > version.ConfigVersion {
		return nil, false, errConfigFromFuture
	}

	if migratedHostMetadata.ConfigVersion == version.ConfigVersion {
		h.Driver = driver
		if err := json.Unmarshal(data, &h); err != nil {
			return nil, migrationPerformed, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
		}
	} else {
		migrationNeeded = true
	}

	if migrationNeeded {
		migrationPerformed = true
		for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion < version.ConfigVersion; h.ConfigVersion++ {
			log.Debugf("Migrating to config v%d", h.ConfigVersion)
			switch h.ConfigVersion {
			case 0:
				hostV0 := &V0{
					Driver: driver,
				}
				if err := json.Unmarshal(data, &hostV0); err != nil {
					return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 0: %s", err)
				}
				hostV1 = MigrateHostV0ToHostV1(hostV0)
			case 1:
				if hostV1 == nil {
					hostV1 = &V1{
						Driver: driver,
					}
					if err := json.Unmarshal(data, &hostV1); err != nil {
						return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 1: %s", err)
					}
				}
				hostV2 = MigrateHostV1ToHostV2(hostV1)
			case 2:
				if hostV2 == nil {
					hostV2 = &V2{
						Driver: driver,
					}
					if err := json.Unmarshal(data, &hostV2); err != nil {
						return nil, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 2: %s", err)
					}
				}
				h = MigrateHostV2ToHostV3(hostV2, data, globalStorePath)
				driver.Data = h.RawDriver
				h.Driver = driver
			case 3:
			}
		}
	}

	h.RawDriver = driver.Data

	return h, migrationPerformed, nil
}
