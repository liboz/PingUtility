package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
)

const PINGER_TIMEOUT_DURATION = time.Duration(250 * time.Millisecond)
const PINGER_WAIT_DURATION = time.Duration(400 * time.Millisecond)
const INTERVAL_BETWEEN_PINGS = time.Duration(500 * time.Millisecond)
const DEFAULT_LOCATION = "Home-PC"
const LOG_FILE_NAME = "log.txt"
const OLD_LOG_FILE_FOLDER = "./old-logs/"
const DEFAULT_OLD_LOG_BASENAME = "log-"

type TargetInfo struct {
	URL       string
	IPAddress *net.IPAddr
}

type PingerReturnInfo struct {
	IterationNumber int
	TargetIndex     int
	Target          string
	PacketsReceived int
	TimeElapsed     time.Duration
}

func runPinger(ch chan PingerReturnInfo, targetIndex int, targetInfo TargetInfo, iterationNumber int, baseId int) {
	pinger, _ := ping.NewPinger(targetInfo.IPAddress.IP.String())
	pinger.SetPrivileged(true)
	pinger.SetNetwork("ip4")
	pinger.Count = 1
	pinger.Timeout = PINGER_TIMEOUT_DURATION
	pinger.SetID(baseId + targetIndex)
	err := pinger.Run()
	info := PingerReturnInfo{IterationNumber: iterationNumber, TargetIndex: targetIndex, Target: targetInfo.URL}
	if err != nil {
		info.PacketsReceived = 0
		select {
		case ch <- info:
			return
		case <-time.After(1 * time.Second):
			fmt.Println("timeout while sending result from ", targetInfo.URL, " on iteration ", iterationNumber)
			return
		}
	}
	stats := pinger.Statistics()

	info.PacketsReceived = stats.PacketsRecv
	if len(stats.Rtts) == 1 {
		info.TimeElapsed = stats.Rtts[0]
	}
	select {
	case ch <- info:
		return
	case <-time.After(1 * time.Second):
		fmt.Println("timeout while sending result from ", targetInfo.URL, " on iteration ", iterationNumber)
	}
}

func formatResult(results []PingerReturnInfo) (string, bool) {
	var sb strings.Builder
	var shouldLog bool
	for index, result := range results {
		sb.WriteString("\"")
		sb.WriteString(result.Target)
		sb.WriteString(": ")
		if result.PacketsReceived == 0 {
			sb.WriteString("TimedOut")
			shouldLog = true
		} else {
			sb.WriteString(strconv.FormatInt(result.TimeElapsed.Milliseconds(), 10))
			sb.WriteString("ms")
		}
		sb.WriteString("\"")
		if index != len(results)-1 {
			sb.WriteString("; ")
		}
	}
	return sb.String(), shouldLog
}

var seed int64 = time.Now().UnixNano()

func getOneHourFromNow() time.Time {
	return time.Now().Add(1 * time.Hour)
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(OLD_LOG_FILE_FOLDER)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Fatalf("error reading folder %s: %v", OLD_LOG_FILE_FOLDER, err)
	}

	result := []string{}
	for _, file := range files {
		result = append(result, file.Name())
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("error encoding json: %v\n", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResult)
	if err != nil {
		fmt.Printf("error writing json: %v\n", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request, filename string) {
	cleanedFileName := strings.ReplaceAll(filename, "..", "")
	http.ServeFile(w, r, OLD_LOG_FILE_FOLDER+cleanedFileName)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	filename, ok := r.URL.Query()["filename"]
	if !ok {
		http.Error(w, "Missing parameter", http.StatusBadRequest)
		return
	}
	cleanedFileName := strings.ReplaceAll(filename[0], "..", "")
	err := os.Remove(OLD_LOG_FILE_FOLDER + cleanedFileName)
	if err != nil {
		fmt.Printf("error deleting file: %v\n", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleGetFilesEndpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		filename, ok := r.URL.Query()["filename"]
		if ok {
			downloadFile(w, r, filename[0])
		} else {
			listFiles(w, r)
		}
		// Serve the resource.
	case http.MethodDelete:
		deleteFile(w, r)
		// Remove the record.
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRequests() {
	http.HandleFunc("/getFiles", handleGetFilesEndpoint)
	log.Fatal(http.ListenAndServe(":7832", nil))
}

func main() {
	f, err := os.OpenFile(LOG_FILE_NAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	targets := [...]string{"facebook.com", "google.com", "localhost"}
	location := DEFAULT_LOCATION

	i := 0
	baseId := rand.New(rand.NewSource(seed)).Intn(math.MaxUint16)
	var targetInfos []TargetInfo

	go handleRequests()

	oneHourFromLastWriteTime := getOneHourFromNow()

	for _, target := range targets {
		ipaddr, err := net.ResolveIPAddr("ip", target)
		if err != nil {
			log.Fatalf("Cannot resolve ip: %v\n", err)
		}
		targetInfo := TargetInfo{URL: target, IPAddress: ipaddr}
		targetInfos = append(targetInfos, targetInfo)
	}

	for {
		start := time.Now()
		ch := make(chan PingerReturnInfo)
		returnedTargets := 0
		for targetIndex, targetInfo := range targetInfos {
			go runPinger(ch, targetIndex, targetInfo, i, baseId)
		}
		results := [len(targets)]PingerReturnInfo{}
	innerLoop:
		for {
			select {
			case res := <-ch:
				returnedTargets += 1
				results[res.TargetIndex] = res
				if returnedTargets == len(targets) {
					break innerLoop
				}
			case <-time.After(PINGER_WAIT_DURATION):
				fmt.Println("timeout 1")
				break innerLoop
			}
		}
		formattedResult, shouldLog := formatResult(results[:])
		formattedTimeStamp := start.Format("2006-01-02 15:04:05.000")
		logString := fmt.Sprintf("%s: %s: [|%s|]\n", formattedTimeStamp, location, formattedResult)
		fmt.Print(logString)

		if shouldLog {
			_, err = f.WriteString(logString)
			if err != nil {
				fmt.Printf("Error writing to file %v\n", err)
			}
		}
		if time.Now().After(oneHourFromLastWriteTime) {
			f.Close()
			newFileName := OLD_LOG_FILE_FOLDER + DEFAULT_OLD_LOG_BASENAME + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".txt"
			err := os.Rename(LOG_FILE_NAME, newFileName)
			if err != nil {
				log.Fatalf("error moving file: %v\n", err)
			}

			f, err = os.OpenFile(LOG_FILE_NAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
			if err != nil {
				log.Fatalf("error moving file: %v\n", err)
			}
			oneHourFromLastWriteTime = getOneHourFromNow()
		}

		elapsedTime := time.Since(start)
		time.Sleep(INTERVAL_BETWEEN_PINGS - elapsedTime)
		i += 1
	}
}
