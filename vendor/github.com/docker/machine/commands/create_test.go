package commands

import (
	"testing"

	"flag"
	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/stretchr/testify/assert"
)

func TestValidateSwarmDiscoveryErrorsGivenInvalidURL(t *testing.T) {
	err := validateSwarmDiscovery("foo")
	assert.Error(t, err)
}

func TestValidateSwarmDiscoveryAcceptsEmptyString(t *testing.T) {
	err := validateSwarmDiscovery("")
	assert.NoError(t, err)
}

func TestValidateSwarmDiscoveryAcceptsValidFormat(t *testing.T) {
	err := validateSwarmDiscovery("token://deadbeefcafe")
	assert.NoError(t, err)
}

type fakeFlagGetter struct {
	flag.Value
	value interface{}
}

func (ff fakeFlagGetter) Get() interface{} {
	return ff.value
}

var nilStringSlice []string

var getDriverOptsFlags = []mcnflag.Flag{
	mcnflag.BoolFlag{
		Name: "bool",
	},
	mcnflag.IntFlag{
		Name: "int",
	},
	mcnflag.IntFlag{
		Name:  "int_defaulted",
		Value: 42,
	},
	mcnflag.StringFlag{
		Name: "string",
	},
	mcnflag.StringFlag{
		Name:  "string_defaulted",
		Value: "bob",
	},
	mcnflag.StringSliceFlag{
		Name: "stringslice",
	},
	mcnflag.StringSliceFlag{
		Name:  "stringslice_defaulted",
		Value: []string{"joe"},
	},
}

var getDriverOptsTests = []struct {
	data     map[string]interface{}
	expected map[string]interface{}
}{
	{
		expected: map[string]interface{}{
			"bool":                  false,
			"int":                   0,
			"int_defaulted":         42,
			"string":                "",
			"string_defaulted":      "bob",
			"stringslice":           nilStringSlice,
			"stringslice_defaulted": []string{"joe"},
		},
	},
	{
		data: map[string]interface{}{
			"bool":             fakeFlagGetter{value: true},
			"int":              fakeFlagGetter{value: 42},
			"int_defaulted":    fakeFlagGetter{value: 37},
			"string":           fakeFlagGetter{value: "jake"},
			"string_defaulted": fakeFlagGetter{value: "george"},
			// NB: StringSlices are not flag.Getters.
			"stringslice":           []string{"ford"},
			"stringslice_defaulted": []string{"zaphod", "arthur"},
		},
		expected: map[string]interface{}{
			"bool":                  true,
			"int":                   42,
			"int_defaulted":         37,
			"string":                "jake",
			"string_defaulted":      "george",
			"stringslice":           []string{"ford"},
			"stringslice_defaulted": []string{"zaphod", "arthur"},
		},
	},
}

func TestGetDriverOpts(t *testing.T) {
	for _, tt := range getDriverOptsTests {
		commandLine := &commandstest.FakeCommandLine{
			LocalFlags: &commandstest.FakeFlagger{
				Data: tt.data,
			},
		}
		driverOpts := getDriverOpts(commandLine, getDriverOptsFlags)
		assert.Equal(t, tt.expected["bool"], driverOpts.Bool("bool"))
		assert.Equal(t, tt.expected["int"], driverOpts.Int("int"))
		assert.Equal(t, tt.expected["int_defaulted"], driverOpts.Int("int_defaulted"))
		assert.Equal(t, tt.expected["string"], driverOpts.String("string"))
		assert.Equal(t, tt.expected["string_defaulted"], driverOpts.String("string_defaulted"))
		assert.Equal(t, tt.expected["stringslice"], driverOpts.StringSlice("stringslice"))
		assert.Equal(t, tt.expected["stringslice_defaulted"], driverOpts.StringSlice("stringslice_defaulted"))
	}
}
