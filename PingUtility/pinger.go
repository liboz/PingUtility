package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
)

const PINGER_TIMEOUT_DURATION = time.Duration(250 * time.Millisecond)
const PINGER_WAIT_DURATION = time.Duration(400 * time.Millisecond)
const INTERVAL_BETWEEN_PINGS = time.Duration(500 * time.Millisecond)
const LOG_FILE_NAME = "log.txt"
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

func loopPinger(targets []string, location string) {
	f, err := os.OpenFile(LOG_FILE_NAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	i := 0
	baseId := rand.New(rand.NewSource(seed)).Intn(math.MaxUint16)
	var targetInfos []TargetInfo

	oneHourFromLastWriteTime := getOneHourFromNow()
	entriesAddedSinceLastFileWrite := 0

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
		results := make([]PingerReturnInfo, len(targets))
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
			} else {
				entriesAddedSinceLastFileWrite += 1
			}
		}
		if entriesAddedSinceLastFileWrite > 20 || time.Now().After(oneHourFromLastWriteTime) {
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
			entriesAddedSinceLastFileWrite = 0
		}

		elapsedTime := time.Since(start)
		time.Sleep(INTERVAL_BETWEEN_PINGS - elapsedTime)
		i += 1
	}
}
