package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	model "specialagentclient/src/model"
	util "specialagentclient/src/util"
	"strings"
	"syscall"
	"time"

	prompt "github.com/c-bata/go-prompt"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

var (
	red              = "\033[31m"
	green            = "\033[32m"
	blue             = "\033[34m"
	light_blue       = "\033[94m"
	reset            = "\033[0m"
	conn             *websocket.Conn
	sendChannel      chan model.ChatDto
	username         string
	serverip         string
	spyname          string
	currentAgentChat string
)

func main() {
	util.ClearScreen()
	// Initialize the sendChannel
	sendChannel = make(chan model.ChatDto, 10)
	fmt.Println("PID:", os.Getpid())
	checkUsername()

	go startWebSocketClient()

	notifyConnect()
	mainMenu(true)

	// Wait for an interrupt signal (e.g., Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

}

func mainMenu(displayBanner bool) {
	if displayBanner {
		myFigure := figure.NewColorFigure("Agent Chat", "big", "green", true)
		myFigure.Print()
		fmt.Println(green + "=========================================================================")
		fmt.Print("Welcome to the special agent communication protocol ")
		fmt.Print(light_blue, username+"(Alias - "+spyname+")\n"+reset)
		util.PrintFlippedText("Please choose one of the following menu items:")
		util.PrintFlippedText("1) Display Online Agents")
		util.PrintFlippedText("2) Start secure text chat")
		util.PrintFlippedText("3) Start secure voice chat")
		util.PrintFlippedText("4) Change Username")

	}

	reader := bufio.NewReader(os.Stdin)
	num, _ := reader.ReadString('\n')

	handleMenu(num)

}

func handleMenu(num string) {
	num = strings.TrimSuffix(num, "\n") // Trim newline character
	num = strings.TrimSuffix(num, "\r") // Trim newline character

	switch num {
	case "1":
		displayAgents()
	case "2":
		util.ClearScreen()
		util.ClearScreen()
		util.ClearScreen()
		promptAgentName()
	case "3":
		handleVoiceChat()
	case "4":
		util.DeleteUsernameFile()
		checkUsername()
	default:
		util.ClearScreen()
		util.PrintFlippedText(red + "Invalid Option - try again" + green)
		mainMenu(true)
	}
}

func handleVoiceChat(){
	util.PrintFlippedText("Enter the Agent name you want to chat with:")
	reader := bufio.NewReader(os.Stdin)
	currentAgentChat, _ = reader.ReadString('\n')
	currentAgentChat = strings.TrimSuffix(currentAgentChat, "\n") // Trim newline character
	currentAgentChat = strings.TrimSuffix(currentAgentChat, "\r") // Trim newline character
	util.ClearScreen()
	fmt.Print(green + "==========================You are chatting with " + blue)
	fmt.Print(green + currentAgentChat)
	fmt.Println(green + "================================")
	fmt.Println(green + "q to quit")

	// Start capturing and encoding audio
	go captureAndEncodeAudio(conn)
}

func captureAndEncodeAudio(c *websocket.Conn) {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize PortAudio:", err)
	}
	defer portaudio.Terminate()

	// Open the default audio input stream
	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, frameSize, func(in []int32) {
		// Convert int32 samples to int16
		samples := make([]int16, len(in))
		for i, v := range in {
			samples[i] = int16(v)
		}

		// Encode the audio samples
		encodedData := encodeAudioSamples(samples)

		// Send the encoded audio data over WebSocket
		err := c.WriteMessage(websocket.BinaryMessage, encodedData)
		if err != nil {
			log.Println("Failed to send audio data over WebSocket:", err)
			return
		}

		// Print the number of samples and data size (for demonstration purposes)
		fmt.Printf("Sent %d audio samples (%d bytes)\n", len(samples), len(encodedData))
	})
	if err != nil {
		log.Fatal("Failed to open audio stream:", err)
	}
	defer stream.Close()

	// Start the audio stream
	err = stream.Start()
	if err != nil {
		log.Fatal("Failed to start audio stream:", err)
	}
	defer stream.Stop()

	// Wait for the stop signal
	<-stopSignal
}

func encodeAudioSamples(samples []int16) []byte {
	// Convert int16 samples to bytes
	buffer := make([]byte, len(samples)*2)
	for i, sample := range samples {
		buffer[i*2] = byte(sample)
		buffer[i*2+1] = byte(sample >> 8)
	}
	return buffer
}

func notifyConnect() {
	//notify server we have connected:
	chat := model.ChatDto{
		From:      spyname,
		To:        "Server",
		Msg:       "",
		Type:      "newConnection",
		Timestamp: time.Now().Unix(),
	}
	sendChannel <- chat
}

func displayAgents() {
	//notify server we have connected:
	chat := model.ChatDto{
		From:      spyname,
		To:        "Server",
		Msg:       "",
		Type:      "whoisonline",
		Timestamp: time.Now().Unix(),
	}
	sendChannel <- chat
}

func promptAgentName() {
	util.PrintFlippedText("Enter the Agent name you want to chat with:")
	completer := func(d prompt.Document) []prompt.Suggest {

		suggestions := []prompt.Suggest{}
		//time.Sleep(1 * time.Second)
		for _, user := range users {
			suggestions = append(suggestions, prompt.Suggest{Text: user})
		}
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
	}

	p := prompt.New(
		func(input string) {
			currentAgentChat = strings.TrimSuffix(input, "\n")            // Trim newline character
			currentAgentChat = strings.TrimSuffix(currentAgentChat, "\r") // Trim newline character
			util.ClearScreen()
			fmt.Print(green + "==========================You are chatting with " + blue)
			fmt.Print(green + currentAgentChat)
			fmt.Println(green + "================================")
			fmt.Println(green + "q to quit")
			handleChat()
		},
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionLivePrefix(func() (string, bool) {
			return ">>> ", true
		}),
	)
	p.Run()
}

func handleChat() {
	reader := bufio.NewReader(os.Stdin)

	for {
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")
		msg = strings.TrimSuffix(msg, "\r")

		if msg == "q" {
			util.ClearScreen()
			mainMenu(true)
			break
		}

		chat := model.ChatDto{
			From:      spyname,
			To:        currentAgentChat,
			Msg:       msg,
			Type:      "chat",
			Timestamp: time.Now().Unix(),
		}

		sendChannel <- chat
	}
}

func checkUsername() {
	spyname = util.GetRandomName()
	// Read the stored username from the file
	u, err := util.ReadUsernameFromFile()
	if err != nil {
		util.ClearScreen()
		strings.Split(u, "") //just ignore this
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your name: ")

		username, _ = reader.ReadString('\n')
		username = strings.TrimSuffix(username, "\n") // Trim newline character
		username = strings.TrimSuffix(username, "\r") // Trim newline character
		//username = username[:len(username)-1] // remove newline character

		// Store the username in a file
		util.StoreUsernameToFile(username)
	} else {
		username = u
	}

	server, err := util.ReadServerIpFromFile()
	if err != nil {
		strings.Split(server, "") //just ignore this
		reader2 := bufio.NewReader(os.Stdin)
		fmt.Print("Enter server IP: ")

		serverip, _ = reader2.ReadString('\n')
		serverip = strings.TrimSuffix(serverip, "\n") // Trim newline character
		serverip = strings.TrimSuffix(serverip, "\r") // Trim newline character
		//serverip = serverip[:len(serverip)-1] // remove newline character
		util.StoreServerIpToFile(serverip)
		util.ClearScreen()
	} else {
		serverip = server
	}
	util.ClearScreen()

}

func startWebSocketClient() {

	err := connectWebSocket("ws://" + serverip + ":8080/ws")
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message:", err)
				return
			}
			//log.Printf("Received message: %s", message)
			chat := model.ChatDto{}
			err = json.Unmarshal(message, &chat)
			if err != nil {
				log.Println("Failed to deserialize message:", err)
				continue
			}
			go receiveMessage(chat)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case message := <-sendChannel:
			err := sendMessage(message)
			if err != nil {
				log.Println("Failed to send message:", err)
				return
			}
		}
	}
}

func connectWebSocket(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	// Add the username as a query parameter to the URL
	q := u.Query()
	q.Set("username", spyname)
	u.RawQuery = q.Encode()

	conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	return nil
}

func receiveMessage(message model.ChatDto) {
	switch strings.TrimSpace(message.Type) {
	case "newConnection":
		fmt.Println(green + "=========================================================================")
		fmt.Println(green + "Connected to the secret network successfully")
		fmt.Println(green + "=========================================================================\n")
	case "whoisonline":
		//fmt.Print(light_blue)
		fmt.Println(message.Users)
		users = message.Users
		reader := bufio.NewReader(os.Stdin)
		num, _ := reader.ReadString('\n')
		num = strings.TrimSuffix(num, "\n") // Trim newline character
		handleMenu(num)
		//fmt.Print(reset)1
	case "chat":
		//fmt.Print(light_blue)
		//fmt.Println(message)
		util.PrintFlippedText(message.From + ":" + message.Msg)
	default:
		// Code to execute when none of the above cases match
		fmt.Println(red + "Unknown Message Type: " + message.Type)
	}
}
func sendMessage(message model.ChatDto) error {
	if conn == nil {
		return fmt.Errorf("WebSocket connection is not established")
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	//fmt.Println("sending message: " + string(jsonData))
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}

	return nil
}
