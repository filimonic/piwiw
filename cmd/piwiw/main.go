package main

import (
	"log"
	"os"
	"slices"
	"strings"

	"github.com/filimonic/piwiw/internal/about"
	"github.com/filimonic/piwiw/internal/piwiw"
)

func main() {
	about.PrintBanner()

	exitCode := 0
	showHelp := false

	if slices.ContainsFunc(os.Args, func(arg string) bool {
		return slices.Contains([]string{"--help", "/?", "-h", "/h"}, strings.ToLower(arg))
	}) {
		showHelp = true
	}

	cfg, err := piwiw.LoadConfig()
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
		showHelp = true
		exitCode = 1
	}
	if showHelp {
		about.PrintHelp()
		os.Exit(exitCode)
	} else {
		piwiw.SetConfig(cfg)
		piwiw.Run()
	}

	os.Exit(0)
}
