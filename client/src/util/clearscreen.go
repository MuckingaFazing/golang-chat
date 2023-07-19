package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func ClearScreen() error {
	if os.Stdout == nil {
		return fmt.Errorf("os.Stdout is not available")
	}

	switch runtime.GOOS {
	case "windows":
		// The escape sequence to clear the screen in a Windows console is \033[2J
		_, err := fmt.Fprint(os.Stdout, "\033[2J\033[0;0H")
		if err != nil {
			return fmt.Errorf("failed to clear Windows console: %w", err)
		}
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Println("Error clearing the screen:", err)
		}
	}
	return nil
}
