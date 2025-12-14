// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicemocks

import (
	"os"
)

func TempPEMFile() (string, error) {
	file, err := os.CreateTemp("", "bundle-*.pem")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(TLSBundleCA)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

var (
	// TLSBundleCA ca.crt
	TLSBundleCA = []byte(`-----BEGIN CERTIFICATE-----
MIICiTCCAfKgAwIBAgIJAJ5X1olt05XjMA0GCSqGSIb3DQEBCwUAMDgxCzAJBgNV
BAYTAkdPMQ8wDQYDVQQIEwZHb3BoZXIxGDAWBgNVBAoTD1Rlc3RpbmcgUk9PVCBD
QTAeFw0xNzAzMDkwMDAyMDZaFw0yNzAzMDcwMDAyMDZaMDgxCzAJBgNVBAYTAkdP
MQ8wDQYDVQQIEwZHb3BoZXIxGDAWBgNVBAoTD1Rlc3RpbmcgUk9PVCBDQTCBnzAN
BgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAw/8DN+t9XQR60jx42rsQ2WE2Dx85rb3n
GQxnKZZLNddsT8rDyxJNP18aFalbRbFlyln5fxWxZIblu9Xkm/HRhOpbSimSqo1y
uDx21NVZ1YsOvXpHby71jx3gPrrhSc/t/zikhi++6D/C6m1CiIGuiJ0GBiJxtrub
UBMXT0QtI2ECAwEAAaOBmjCBlzAdBgNVHQ4EFgQU8XG3X/YHBA6T04kdEkq6+4GV
YykwaAYDVR0jBGEwX4AU8XG3X/YHBA6T04kdEkq6+4GVYymhPKQ6MDgxCzAJBgNV
BAYTAkdPMQ8wDQYDVQQIEwZHb3BoZXIxGDAWBgNVBAoTD1Rlc3RpbmcgUk9PVCBD
QYIJAJ5X1olt05XjMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADgYEAeILv
z49+uxmPcfOZzonuOloRcpdvyjiXblYxbzz6ch8GsE7Q886FTZbvwbgLhzdwSVgG
G8WHkodDUsymVepdqAamS3f8PdCUk8xIk9mop8LgaB9Ns0/TssxDvMr3sOD2Grb3
xyWymTWMcj6uCiEBKtnUp4rPiefcvCRYZ17/hLE=
-----END CERTIFICATE-----
`)
)
