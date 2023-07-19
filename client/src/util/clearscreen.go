package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func ClearScreen() {
	clearCommand := ""
	switch runtime.GOOS {
	case "windows":
		clearCommand = "cls"
	default:
		clearCommand = "clear"
	}

	cmd := exec.Command(clearCommand)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Error clearing the screen:", err)
		return
	}
}