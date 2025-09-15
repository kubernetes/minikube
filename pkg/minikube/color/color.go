package color

// colorFunc allows generic coloring
type colorFunc func(string, ...interface{}) string

func isEmoji(s string) bool {
	if s == "✅" {
		return true
	}
	return false
}

func ColorRow(row []string, colored colorFunc) {
	for i := range row {
		if row[i] == "" || isEmoji(row[i]) {
			continue
		}
		row[i] = colored("%s", row[i])
	}
}