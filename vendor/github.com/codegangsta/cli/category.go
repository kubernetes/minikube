package cli

type CommandCategories []*CommandCategory

type CommandCategory struct {
	Name     string
	Commands Commands
}

func (c CommandCategories) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

func (c CommandCategories) Len() int {
	return len(c)
}

func (c CommandCategories) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CommandCategories) AddCommand(category string, command Command) CommandCategories {
	for _, commandCategory := range c {
		if commandCategory.Name == category {
			commandCategory.Commands = append(commandCategory.Commands, command)
			return c
		}
	}
	return append(c, &CommandCategory{Name: category, Commands: []Command{command}})
}
