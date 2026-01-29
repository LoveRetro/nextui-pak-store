//go:build ignore

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/LoveRetro/nextui-pak-store/models"
)

type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Encoding    string `json:"encoding"`
	Content     string `json:"content"`
	DownloadUrl string `json:"download_url"`
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func logHeader(msg string) {
	fmt.Printf("\n%s%s==> %s%s\n", colorBold, colorCyan, msg, colorReset)
}

func logSuccess(msg string) {
	fmt.Printf("  %s✓%s %s\n", colorGreen, colorReset, msg)
}

func logSkipped(msg string) {
	fmt.Printf("  %s○%s %s %s(disabled)%s\n", colorYellow, colorReset, msg, colorYellow, colorReset)
}

func logError(msg string) {
	fmt.Printf("  %s✗%s %s\n", colorRed, colorReset, msg)
}

func logFatal(msg string) {
	fmt.Printf("\n%s%s✗ Error: %s%s\n", colorBold, colorRed, msg, colorReset)
	os.Exit(1)
}

func main() {
	fmt.Printf("%s%sNextUI Pak Store - Storefront Builder%s\n", colorBold, colorCyan, colorReset)

	logHeader("Loading storefront_base.json")
	data, err := os.ReadFile("storefront_base.json")
	if err != nil {
		logFatal("Error reading file: " + err.Error())
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		logFatal("Unable to unmarshal storefront: " + err.Error())
	}
	logSuccess(fmt.Sprintf("Found %d paks", len(sf.Paks)))

	logHeader("Fetching pak data from GitHub")

	var paks []models.Pak
	successCount := 0
	skippedCount := 0

	for _, p := range sf.Paks {
		repoPath := strings.ReplaceAll(p.RepoURL, models.GitHubRoot, "")
		parts := strings.Split(repoPath, "/")
		if len(parts) < 2 {
			logFatal("Invalid repository URL format: " + p.RepoURL)
		}

		owner := parts[0]
		repo := parts[1]

		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
			owner, repo, models.PakJsonStub)

		pak := models.Pak{}

		if p.Disabled {
			logSkipped(p.StorefrontName + " | " + p.RepoURL)
			skippedCount++
		} else {
			pak, err = fetchPakJsonFromGitHubAPI(apiURL)
			if err != nil {
				logError(fmt.Sprintf("%s | %s - %v", p.StorefrontName, p.RepoURL, err))
				logFatal("Unable to fetch pak json for " + p.StorefrontName + " (" + p.RepoURL + ")")
			}
			logSuccess(p.StorefrontName + " | " + p.RepoURL)
			successCount++
		}

		pak.ID = p.ID
		pak.StorefrontName = p.StorefrontName
		pak.PreviousNames = p.PreviousNames
		pak.PreviousRepoURLs = p.PreviousRepoURLs
		pak.RepoURL = p.RepoURL
		pak.Categories = p.Categories
		pak.LargePak = p.LargePak
		pak.Disabled = p.Disabled

		paks = append(paks, pak)
	}

	sf.Paks = paks

	logHeader("Writing storefront.json")
	jsonData, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		logFatal("Unable to marshal storefront to JSON: " + err.Error())
	}

	err = os.WriteFile("storefront.json", jsonData, 0644)
	if err != nil {
		logFatal("Unable to write storefront.json: " + err.Error())
	}

	fmt.Printf("\n%s%s✓ Complete!%s %d paks fetched", colorBold, colorGreen, colorReset, successCount)
	if skippedCount > 0 {
		fmt.Printf(", %d skipped", skippedCount)
	}
	fmt.Println()
}

func fetchPakJsonFromGitHubAPI(apiURL string) (models.Pak, error) {
	var pak models.Pak

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return pak, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")

	req.Header.Add("Authorization", "Bearer "+os.Getenv("GH_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return pak, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return pak, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var content GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return pak, fmt.Errorf("error decoding GitHub API response: %w", err)
	}

	if content.Encoding == "base64" {
		contentBytes, err := base64.StdEncoding.DecodeString(
			strings.ReplaceAll(content.Content, "\n", ""))
		if err != nil {
			return pak, fmt.Errorf("error decoding base64 content: %w", err)
		}

		if err := json.Unmarshal(contentBytes, &pak); err != nil {
			return pak, fmt.Errorf("error parsing pak.json: %w", err)
		}
	} else {
		return pak, fmt.Errorf("unexpected content encoding: %s", content.Encoding)
	}

	return pak, nil
}
