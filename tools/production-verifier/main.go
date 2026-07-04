package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

type readiness struct {
	Ready  bool            `json:"ready"`
	Checks map[string]bool `json:"checks"`
}

type wallet struct {
	BillingAccountID string `json:"billingAccountId"`
}

type createWorkspaceResult struct {
	WorkspaceID string `json:"workspaceId"`
	URL         string `json:"url"`
	State       string `json:"state"`
	ApprovalID  string `json:"approvalId"`
}

type authSession struct {
	CSRFToken string `json:"csrfToken"`
}

func main() {
	origin := strings.TrimRight(os.Getenv("OPL_CONSOLE_ORIGIN"), "/")
	if origin == "" {
		log.Fatal("OPL_CONSOLE_ORIGIN is required")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{Timeout: 10 * time.Second, Jar: jar}
	production := getReadiness(client, origin+"/api/production/readiness")
	runtime := getReadiness(client, origin+"/api/runtime/readiness")

	result := map[string]any{
		"production": production,
		"runtime":    runtime,
	}

	if !production.Ready || !runtime.Ready {
		_ = json.NewEncoder(os.Stdout).Encode(result)
		log.Fatal("readiness gate failed")
	}
	if os.Getenv("OPL_VERIFY_WORKSPACE") == "1" {
		result["workspace"] = verifyWorkspace(client, origin)
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)
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

func verifyWorkspace(client *http.Client, origin string) createWorkspaceResult {
	email := os.Getenv("OPL_CONSOLE_EMAIL")
	password := os.Getenv("OPL_CONSOLE_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("OPL_CONSOLE_EMAIL and OPL_CONSOLE_PASSWORD are required when OPL_VERIFY_WORKSPACE=1")
	}
	var session authSession
	postJSON(client, origin+"/api/auth/login", map[string]string{"email": email, "password": password}, &session, "")
	if session.CSRFToken == "" {
		log.Fatal("login did not return csrfToken")
	}

	var account wallet
	getJSON(client, origin+"/api/billing/wallet", &account)
	if account.BillingAccountID == "" {
		log.Fatal("billing wallet did not return billingAccountId")
	}

	workspaceID := "verify-" + time.Now().UTC().Format("20060102150405")
	token := "verify-" + time.Now().UTC().Format("150405")
	var created createWorkspaceResult
	postJSON(client, origin+"/api/workspaces", map[string]string{
		"workspaceId":      workspaceID,
		"name":             "Production Verifier Workspace",
		"billingAccountId": account.BillingAccountID,
		"packageId":        "basic",
		"token":            token,
	}, &created, session.CSRFToken)
	if created.State == "approval_required" {
		log.Fatalf("workspace verifier requires approval id=%s; approve policy or disable smoke lifecycle for this run", created.ApprovalID)
	}
	if created.URL == "" {
		log.Fatalf("workspace verifier returned empty URL: %#v", created)
	}
	handoffURL := created.URL
	if strings.HasPrefix(handoffURL, "/") {
		handoffURL = origin + handoffURL
	}
	var handoff createWorkspaceResult
	getJSON(client, handoffURL, &handoff)
	if handoff.URL == "" {
		log.Fatalf("workspace handoff returned empty URL: %#v", handoff)
	}
	postJSON(client, origin+"/api/workspaces/"+workspaceID+"/delete", map[string]string{}, nil, session.CSRFToken)
	return handoff
}

func getJSON(client *http.Client, url string, target any) {
	response, err := client.Get(url)
	if err != nil {
		log.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("GET %s: status %d", url, response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		log.Fatalf("decode %s: %v", url, err)
	}
}

func postJSON(client *http.Client, url string, payload any, target any, csrfToken string) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("marshal %s: %v", url, err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Fatalf("build POST %s: %v", url, err)
	}
	request.Header.Set("content-type", "application/json")
	if csrfToken != "" {
		request.Header.Set("x-opl-csrf-token", csrfToken)
	}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		log.Fatalf("POST %s: status %d", url, response.StatusCode)
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			log.Fatalf("decode %s: %v", url, err)
		}
	}
}
