// Originally Author is @ailispaw
// import by https://github.com/ailispaw/xhyvectl/tree/master/vmnet
package vmnet

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	CONFIG_PLIST = "/Library/Preferences/SystemConfiguration/com.apple.vmnet"
	NET_ADDR_KEY = "Shared_Net_Address"
	NET_MASK_KEY = "Shared_Net_Mask"
)

func GetNetAddr() (net.IP, error) {
	if !IsExist(CONFIG_PLIST + ".plist") {
		return nil, fmt.Errorf("Does not exist %s", CONFIG_PLIST+".plist")
	}

	out, err := exec.Command("defaults", "read", CONFIG_PLIST, NET_ADDR_KEY).Output()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(strings.TrimSpace(string(out)))
	if ip == nil {
		return nil, fmt.Errorf("Could not get the network address for vmnet")
	}
	return ip, nil
}

func getNetMask() (net.IPMask, error) {
	if !IsExist(CONFIG_PLIST + ".plist") {
		return nil, fmt.Errorf("Does not exist %s", CONFIG_PLIST+".plist")
	}

	out, err := exec.Command("defaults", "read", CONFIG_PLIST, NET_MASK_KEY).Output()
	if err != nil {
		return nil, err
	}
	mask := net.ParseIP(strings.TrimSpace(string(out)))
	if mask == nil {
		return nil, fmt.Errorf("Could not get the network mask for vmnet")
	}
	return net.IPMask(mask.To4()), nil
}

func GetIPNet() (*net.IPNet, error) {
	ip, err := GetNetAddr()
	if err != nil {
		return nil, err
	}

	mask, err := getNetMask()
	if err != nil {
		return nil, err
	}

	return &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}, nil
}

// IsExist returns whether the filename is exists.
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
