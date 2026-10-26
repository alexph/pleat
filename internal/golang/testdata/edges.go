package orders

func directError() error {
	value, err := readValue()
	// This comment remains attached to the error check.
	if err != nil {
		return err
	}

	use(value)

	return nil
}

func smallGuard(ready bool) {
	if !ready {
		return
	}
}

func smallReturn() int {
	value := 1
	return value
}

func multiline() {
	value := combine(
		1,

		2,
	)
	consume(value)
	finish()
}

func clauses() {
	if first() {
		one()
	} else if second() {
		two()
	} else {
		three()
	}

	switch choice() {
	case 1:
		one()
		two()
		return
	case 2:
		three()
	}
}
