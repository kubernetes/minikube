package commands

import "github.com/codegangsta/cli"

type ByFlagName []cli.Flag

func (flags ByFlagName) Len() int {
	return len(flags)
}

func (flags ByFlagName) Swap(i, j int) {
	flags[i], flags[j] = flags[j], flags[i]
}

func (flags ByFlagName) Less(i, j int) bool {
	return flags[i].String() < flags[j].String()
}
