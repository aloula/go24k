//go:build !fyne

package main

import "fmt"

func launchGUI() {
	fmt.Println("GUI support is not included in this build.")
	fmt.Println("Rebuild with Fyne enabled: go build -tags fyne -o go24k .")
}
