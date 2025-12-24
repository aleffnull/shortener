package main

import (
	"log"
	"math/rand"
	"os"
)

func main() {
	x := rand.Int()
	if x%2 == 0 {
		log.Fatal("even")
	}

	os.Exit(0)
}

//lint:ignore U1000 "linter test code"
func osExit() {
	x := rand.Int()
	if x%2 == 0 {
		os.Exit(1) // want "os.Exit is allowed only inside of main.main"
	}

	os.Exit(0) // want "os.Exit is allowed only inside of main.main"
}

//lint:ignore U1000 "linter test code"
func logFatal() {
	x := rand.Int()
	if x%2 == 0 {
		log.Fatal("even") // want "log.Fatal is allowed only inside of main.main"
	}

	log.Fatal("odd") // want "log.Fatal is allowed only inside of main.main"
}
