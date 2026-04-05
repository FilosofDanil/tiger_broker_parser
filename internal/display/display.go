package display

import (
	"fmt"
	"time"
)

func PrintHeader(title string) {
	border := "═══════════════════════════════════════════════════════"
	fmt.Printf("\n%s\n  %s\n  %s\n%s\n\n", border, title, time.Now().UTC().Format("2006-01-02 15:04:05 UTC"), border)
}

func PrintSection(title string) {
	fmt.Printf("┌─ %s\n", title)
}
