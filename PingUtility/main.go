package main

import (
	"os"
)

const DEFAULT_LOCATION = "Home-PC"
const OLD_LOG_FILE_FOLDER = "./old-logs/"

func main() {
	argsWithoutProg := os.Args[1:]
	targets := []string{"facebook.com", "google.com", "localhost"}
	location := DEFAULT_LOCATION
	if len(argsWithoutProg) > 0 {
		location = argsWithoutProg[0]
	}
	if len(argsWithoutProg) > 1 {
		targets[0] = argsWithoutProg[1]
	}

	go handleRequests()
	loopPinger(targets, location)
}
