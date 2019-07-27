package initflag

import (
	"flag"
)

func init() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	_ = fs.Parse([]string{})
	flag.CommandLine = fs
}
