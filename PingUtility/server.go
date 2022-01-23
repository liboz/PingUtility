package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

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
	http.HandleFunc("/logFile", handleGetFilesEndpoint)
	log.Fatal(http.ListenAndServe(":7832", nil))
}
