package extract

import "fmt"

func DoSomeStuff() {
	// Test with a URL
	PrintToScreenNoInterface("http://kubernetes.io")

	// Test with something that Go thinks looks like a URL
	PrintToScreenNoInterface("Hint: This is not a URL, come on.")

	// Try with an integer
	PrintToScreenNoInterface("5")

	// Try with a sudo command
	path := "."
	PrintToScreen("sudo ls %s", path)

	DoSomeOtherStuff(true, 4, "I think this should work")

	v := "This is a variable with a string assigned"
	PrintToScreenNoInterface(v)
}

func DoSomeOtherStuff(choice bool, i int, s string) {

	// Let's try an if statement
	if choice {
		PrintToScreen("This was a choice: %s", s)
	} else if i > 5 {
		PrintToScreen("Wow another string: %s", i)
	} else {
		for i > 10 {
			PrintToScreenNoInterface("Holy cow I'm in a loop!")
			i = i + 1
		}
	}

}

func PrintToScreenNoInterface(s string) {
	PrintToScreen(s, nil)
}

// This will be the function we'll focus the extractor on
func PrintToScreen(s string, i interface{}) {
	fmt.Printf(s, i)
}
