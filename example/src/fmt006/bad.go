package fmt006

func MissingAfterIf(cond bool) {
	if cond {
		return
	}
	doSomething() // want `FMT-006: parent block must contain a blank line after a block ending with return`
}

func MissingAfterLoop(items []string) {
	for _, item := range items {
		if item == "" {
			return
		}
		use(item) // want `FMT-006: parent block must contain a blank line after a block ending with return`
	}
}

func MissingAfterRange(items []string) {
	for _, item := range items {
		if item == "" {
			return
		}
	}
	doSomething() // want `FMT-006: parent block must contain a blank line after a block ending with return`
}

func MissingAfterSwitch(v int) {
	switch v {
	case 0:
		return
	}
	doSomething() // want `FMT-006: parent block must contain a blank line after a block ending with return`
}

func doSomething() {}

func use(_ string) {}
