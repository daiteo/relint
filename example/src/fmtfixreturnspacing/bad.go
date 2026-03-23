package fmtfixreturnspacing

func KeepBlankLines(cond bool, value int) { // want `FMTFIX: apply format fixes \(merge declaration blocks, reorder declarations\)`
	if cond {
		return
	}
	doSomething()

	switch value {
	case 0:
		return
	}
	// keep: comment attached to the following statement
	doSomething()
}

func doSomething() {}
