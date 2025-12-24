package panic_pkg

//lint:ignore U1000 "linter test code"
func callPanic() {
	panic("panic called") // want "panic is prohibited"
}
