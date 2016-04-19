package host

import (
	"bytes"
	"encoding/json"

	"github.com/docker/machine/libmachine/log"
)

type RawHost struct {
	Driver *json.RawMessage
}

func MigrateHostV2ToHostV3(hostV2 *V2, data []byte, storePath string) *Host {
	// Migrate to include RawDriver so that driver plugin will work
	// smoothly.
	rawHost := &RawHost{}
	if err := json.Unmarshal(data, &rawHost); err != nil {
		log.Warnf("Could not unmarshal raw host for RawDriver information: %s", err)
	}

	m := make(map[string]interface{})

	// Must migrate to include store path in driver since it was not
	// previously stored in drivers directly
	d := json.NewDecoder(bytes.NewReader(*rawHost.Driver))
	d.UseNumber()
	if err := d.Decode(&m); err != nil {
		log.Warnf("Could not unmarshal raw host into map[string]interface{}: %s", err)
	}

	m["StorePath"] = storePath

	// Now back to []byte
	rawDriver, err := json.Marshal(m)
	if err != nil {
		log.Warnf("Could not re-marshal raw driver: %s", err)
	}

	h := &Host{
		ConfigVersion: 2,
		DriverName:    hostV2.DriverName,
		Name:          hostV2.Name,
		HostOptions:   hostV2.HostOptions,
		RawDriver:     rawDriver,
	}

	return h
}
