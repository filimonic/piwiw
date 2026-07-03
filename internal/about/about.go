package about

import (
	_ "embed"
	"fmt"

	"github.com/filimonic/piwiw/internal/buildinfo"
)

//go:embed help.txt
var help []byte

func PrintBanner() {
	fmt.Printf("%s %s+%s (built %s)\n", "piwiw", buildinfo.Version, buildinfo.CommitSHA, buildinfo.BuildDate)
}

func PrintHelp() {
	fmt.Print(string(help))
	fmt.Println()
}
