package main

import (
	"os"

	"github.com/angelini/netz/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
