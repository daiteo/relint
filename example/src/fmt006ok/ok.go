package fmt006ok

func GoodAfterIf(cond bool) {
	if cond {
		return
	}

	doSomething()
}

func GoodInsideLoop(items []string) {
	for _, item := range items {
		if item == "" {
			return
		}

		use(item)
	}
}

func GoodAfterSwitch(v int) {
	switch v {
	case 0:
		return
	}

	doSomething()
}

func doSomething() {}

func use(_ string) {}
