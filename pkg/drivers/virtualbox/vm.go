package virtualbox

import "strconv"

type VM struct {
	CPUs   int
	Memory int
}

func getVMInfo(name string, vbox VBoxManager) (*VM, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	vm := &VM{}

	err = parseKeyValues(out, reEqualLine, func(key, val string) error {
		switch key {
		case "cpus":
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			vm.CPUs = v
		case "memory":
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			vm.Memory = v
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return vm, nil
}
