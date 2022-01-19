package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

const configPath = "./remote_config.json"
const remoteLogsFolder = "./RemoteLogs/"
const processedLogsFolder = "./Processed/"

const insertTimeoutDataSQL = `INSERT INTO timeout_data (name, location, timestamp, hour_minute) VALUES (?, ?, ?, ?)`
const timeStampDataLayout = "2006-01-02 15:04:05.000"
const hourMinuteLayout = "15:04"

type RemoteConfig struct {
	Targets        []Target `json:"targets"`
	BackupLocation string   `json:"backuplocation"`
}

type Target struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type LogFile struct {
	FileName   string
	URL        string
	RemoteName string
	LocalName  string
}

type LogData struct {
	Timestamp string
	Location  string
	Name      string
}

func parseRemoteConfig() RemoteConfig {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		log.Println("Config file does not exist")
		time.Sleep(60 * time.Second)
		os.Exit(1)
	}

	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var remoteConfig RemoteConfig

	json.Unmarshal(byteValue, &remoteConfig)

	return remoteConfig
}

func downloadTextFile(logFile LogFile) error {
	res, err := http.Get(logFile.URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(logFile.LocalName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func deleteTextFile(logFile LogFile) error {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, logFile.URL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	return err
}

func getTextFiles(remoteConfig RemoteConfig) []LogFile {
	files := []LogFile{}

	for _, target := range remoteConfig.Targets {
		res, err := http.Get(target.URL)
		if err != nil {
			log.Println(err)
			continue
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}

		logFilesOnServer := []string{}
		err = json.Unmarshal(body, &logFilesOnServer)
		if err != nil {
			log.Println(err)
		}

		for _, logFile := range logFilesOnServer {
			fileData := LogFile{FileName: logFile,
				URL:        target.URL + "/" + logFile,
				RemoteName: target.Name,
				LocalName:  target.Name + "-" + logFile}
			files = append(files, fileData)
			downloadTextFile(fileData)
		}
		if len(logFilesOnServer) > 0 {
			log.Printf("Downloaded %d files from %s\n", len(logFilesOnServer), target.Name)
		}
	}

	return files
}

func insertIntoSqlLite(db *sql.DB, logData LogData) {
	parsedTime, err := time.Parse(timeStampDataLayout, logData.Timestamp)
	if err != nil {
		log.Println(err.Error())
		return
	}

	hourMinute := parsedTime.Format(hourMinuteLayout)
	statement, err := db.Prepare(insertTimeoutDataSQL) // Prepare statement.
	// avoid SQL injections
	if err != nil {
		log.Println(err.Error())
		return
	}
	_, err = statement.Exec(logData.Name, logData.Location, logData.Timestamp, hourMinute)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func parseLogAndInsertIntoSqlLite(logFile LogFile, r *regexp.Regexp) {
	db, _ := sql.Open("sqlite3", "./data.db")
	defer db.Close()

	file, err := os.Open(logFile.LocalName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	entriesAdded := 0

	for fileScanner.Scan() {
		line := r.ReplaceAllString(fileScanner.Text(), "")
		parts := strings.Split(line, ": ")
		var allParts []string
		var data []LogData
		for _, part := range parts {
			if strings.Contains(part, "; ") {
				allParts = append(allParts, strings.Split(part, "; ")...)
			} else {
				allParts = append(allParts, part)
			}
		}

		var timestamp string
		var location string
		var name string

		for index, part := range allParts {
			if index == 0 {
				timestamp = part
			} else if index == 1 {
				location = part
			} else if index%2 == 0 {
				name = part
			} else {
				if part == "TimedOut" {
					data = append(data, LogData{Timestamp: timestamp, Location: location, Name: name})
				}
			}
		}

		for _, dataItem := range data {
			insertIntoSqlLite(db, dataItem)
			entriesAdded += 1
		}
	}

	newPath := processedLogsFolder + logFile.LocalName
	err = os.Rename(logFile.LocalName, newPath)
	log.Printf("finished processing %s and moving to %s. Added %d new entries. Deleting on server.\n", logFile.LocalName, newPath, entriesAdded)
	deleteTextFile(logFile)
	log.Printf("Deleted %s on server %s.\n", logFile.FileName, logFile.RemoteName)
}

func main() {
	r := regexp.MustCompile("[[\\]|\"\n]")

	remoteConfig := parseRemoteConfig()
	fmt.Println(remoteConfig)
	i := 0
	for {
		currTime := time.Now().Format("2006-01-02 15:04:05")
		textFiles := getTextFiles(remoteConfig)
		if len(textFiles) == 0 {
			log.Printf("%s: found no new log files\n", currTime)
		} else {
			for _, textFile := range textFiles {
				parseLogAndInsertIntoSqlLite(textFile, r)
			}
		}

		time.Sleep(1 * time.Minute)
		i += 1
		if i%60 == 0 {
			log.Printf("%s: uploading db to remote backup\n", currTime)
			_, err := exec.Command("scp", "data.db", fmt.Sprintf("%s:/root/", remoteConfig.BackupLocation)).Output()
			if err != nil {
				log.Println(err)
			}
		}
	}
}
