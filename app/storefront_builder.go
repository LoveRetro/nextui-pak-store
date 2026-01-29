//go:build ignore

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func main() {
	data, err := os.ReadFile("storefront_base.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		log.Fatal("Unable to unmarshal storefront", err)
	}

	var paks []models.Pak

	for _, p := range sf.Paks {
		repoPath := strings.ReplaceAll(p.RepoURL, models.GitHubRoot, "")
		parts := strings.Split(repoPath, "/")
		if len(parts) < 2 {
			log.Fatal("Invalid repository URL format:", p.RepoURL)
		}

		owner := parts[0]
		repo := parts[1]

		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
			owner, repo, models.PakJsonStub)

		pak := models.Pak{}
		var err error

		if !p.Disabled {
			pak, err = fetchPakJsonFromGitHubAPI(apiURL)
			if err != nil {
				log.Fatal("Unable to fetch pak json for "+p.Name+" ("+p.RepoURL+")", err)
			}
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

	jsonData, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		log.Fatal("Unable to marshal storefront to JSON", err)
	}

	err = os.WriteFile("storefront.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Unable to write storefront.json", err)
	}
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
