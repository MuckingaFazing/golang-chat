package util

import (
	"bufio"
	"os"
)

func StoreServerIpToFile(ip string) error {
	file, err := os.Create("server.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(ip)
	if err != nil {
		return err
	}

	return nil
}

func StoreUsernameToFile(username string) error {
	file, err := os.Create("username.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(username)
	if err != nil {
		return err
	}

	return nil
}

func ReadServerIpFromFile() (string, error) {
	file, err := os.Open("server.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	server := scanner.Text()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return server, nil
}

func ReadUsernameFromFile() (string, error) {
	file, err := os.Open("username.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	username := scanner.Text()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return username, nil
}

func DeleteUsernameFile() error {
	err := os.Remove("username.txt")
	if err != nil {
		return err
	}

	return nil
}
