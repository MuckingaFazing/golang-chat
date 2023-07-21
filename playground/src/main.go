package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// Constants for audio settings
const (
	SampleRate   = 44100
	NumChannels  = 1
	SampleFormat = 16
	BufferSize   = 4096
)

func main() {
	// Initialize audio capture
	captureAudio()

	// Wait for an interrupt signal (e.g., Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt
}

func captureAudio() {
	// Open the audio device for capturing
	audioDev, err := syscall.Open("/dev/dsp", syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Println("Failed to open audio device:", err)
		return
	}
	defer syscall.Close(audioDev)

	// Set audio parameters
	param := &syscall.AudioParams{
		Freq:       SampleRate,
		Format:     syscall.AFMT_S16_LE,
		Channels:   NumChannels,
		Access:     syscall.O_RDONLY,
		NonBlocking: 1,
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(audioDev), syscall.SNDCTL_DSP_SETFMT, uintptr(unsafe.Pointer(param)))
	if errno != 0 {
		fmt.Println("Failed to set audio format:", errno)
		return
	}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(audioDev), syscall.SNDCTL_DSP_CHANNELS, uintptr(unsafe.Pointer(param)))
	if errno != 0 {
		fmt.Println("Failed to set audio channels:", errno)
		return
	}

	// Alternative method for setting the sample rate
	err = setAudioSampleRate(audioDev, SampleRate)
	if err != nil {
		fmt.Println("Failed to set audio sample rate:", err)
		return
	}

	// Create a buffer to store audio data
	buffer := make([]byte, BufferSize)

	// Start capturing audio
	fmt.Println("Capturing audio...")
	for {
		// Read audio data into the buffer
		_, err := syscall.Read(audioDev, buffer)
		if err != nil {
			fmt.Println("Failed to read audio data:", err)
			return
		}

		// Process the captured audio data here
		// You can send it over a WebSocket, perform audio analysis, etc.
		// For simplicity, this example just prints the size of the captured data.
		fmt.Println("Captured audio data size:", len(buffer))
	}
}

func setAudioSampleRate(audioDev int, sampleRate int) error {
	// Define the constant specific to your platform for setting the sample rate
	const ioctlSetSampleRate = 0x8004620B // Example constant value, replace with the correct one for your platform

	// Set the sample rate using the ioctl system call
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(audioDev), uintptr(ioctlSetSampleRate), uintptr(sampleRate))
	if errno != 0 {
		return errno
	}

	return nil
}
