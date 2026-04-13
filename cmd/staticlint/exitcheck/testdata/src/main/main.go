package main

import "os"

func main() {
	os.Exit(1) // want "direct call to os.Exit in main function of package main is not allowed"
}

func other() {
	os.Exit(0) // ok: not in main function
}
