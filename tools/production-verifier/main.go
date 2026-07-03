package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type readiness struct {
	Ready  bool            `json:"ready"`
	Checks map[string]bool `json:"checks"`
}

func main() {
	origin := strings.TrimRight(os.Getenv("OPL_CONSOLE_ORIGIN"), "/")
	if origin == "" {
		log.Fatal("OPL_CONSOLE_ORIGIN is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	production := getReadiness(client, origin+"/api/production/readiness")
	runtime := getReadiness(client, origin+"/api/runtime/readiness")

	result := map[string]readiness{
		"production": production,
		"runtime":    runtime,
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)

	if !production.Ready || !runtime.Ready {
		log.Fatal("readiness gate failed")
	}
}

func getReadiness(client *http.Client, url string) readiness {
	response, err := client.Get(url)
	if err != nil {
		log.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("GET %s: status %d", url, response.StatusCode)
	}
	var report readiness
	if err := json.NewDecoder(response.Body).Decode(&report); err != nil {
		log.Fatalf("decode %s: %v", url, err)
	}
	fmt.Fprintf(os.Stderr, "%s ready=%v checks=%v\n", url, report.Ready, report.Checks)
	return report
}
