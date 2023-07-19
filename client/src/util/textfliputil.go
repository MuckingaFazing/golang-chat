package util

import (
	"fmt"
	"math/rand"
	"time"
)

var delay = 5
var numFlips = 3

var (
	red   = "\033[31m"
	green = "\033[32m"
	blue  = "\033[34m"
	reset = "\033[0m"
)

func PrintFlippedText(text string) {
	for _, ch := range text {
		flipped := FlipCharacterMultipleTimes(ch, numFlips)
		fmt.Print(green + string(flipped))
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	for i := len(text) - 1; i >= 0; i-- {
		fmt.Print("\b \b") // Erase the previously printed character
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	for _, ch := range text {
		fmt.Print(string(ch))
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	fmt.Println()
}

func FlipCharacterMultipleTimes(ch rune, numFlips int) rune {
	flipped := ch
	for i := 0; i < numFlips; i++ {
		if rand.Intn(2) == 0 { // Randomly decide whether to flip the character
			flipped = FlipCharacter(flipped)
		}
	}
	return flipped
}

func FlipCharacter(ch rune) rune {
	switch {
	case ch >= 'a' && ch <= 'z':
		return 'a' + 'z' - ch
	case ch >= 'A' && ch <= 'Z':
		return 'A' + 'Z' - ch
	default:
		return ch
	}
}