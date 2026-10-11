package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/LoveRetro/nextui-pak-store/models"
)

type Result struct {
	Pak        models.Pak
	URL        string
	StatusCode int
	Error      error
}

func (r Result) failed() bool {
	return r.Error != nil || r.StatusCode != http.StatusOK
}

func (r Result) transient() bool {
	return r.Error != nil || r.StatusCode == http.StatusTooManyRequests || r.StatusCode >= 500
}

func (r Result) label() string {
	if r.Pak.Experimental {
		return " [experimental]"
	}
	return ""
}

func main() {
	storefrontPath := "storefront.json"
	if len(os.Args) > 1 {
		storefrontPath = os.Args[1]
	}

	data, err := os.ReadFile(storefrontPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", storefrontPath, err)
		os.Exit(1)
	}

	var storefront models.Storefront
	if err := json.Unmarshal(data, &storefront); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing storefront JSON: %v\n", err)
		os.Exit(1)
	}

	var activePaks []models.Pak
	for _, p := range storefront.Paks {
		if !p.Disabled {
			p.Experimental = false
			activePaks = append(activePaks, p)
		}
	}
	for _, p := range storefront.ExperimentalPaks {
		if !p.Disabled {
			p.Experimental = true
			activePaks = append(activePaks, p)
		}
	}

	fmt.Printf("Verifying release artifacts for %d paks...\n\n", len(activePaks))

	client := &http.Client{Timeout: 15 * time.Second}

	results := make([]Result, len(activePaks))
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for i, pak := range activePaks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = verify(client, pak)
		}()
	}

	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Pak.StorefrontName < results[j].Pak.StorefrontName
	})

	var passed, failed []Result
	for _, r := range results {
		if r.failed() {
			failed = append(failed, r)
		} else {
			passed = append(passed, r)
		}
	}

	for _, r := range passed {
		fmt.Printf("  PASS  %s (%s)%s\n", r.Pak.StorefrontName, r.Pak.Version, r.label())
	}

	if len(failed) > 0 {
		fmt.Printf("\n--- FAILURES ---\n\n")
		for _, r := range failed {
			if r.Error != nil {
				fmt.Printf("  FAIL  %s (%s)%s\n        Error: %v\n        URL: %s\n\n", r.Pak.StorefrontName, r.Pak.Version, r.label(), r.Error, r.URL)
			} else {
				fmt.Printf("  FAIL  %s (%s)%s\n        HTTP %d\n        URL: %s\n\n", r.Pak.StorefrontName, r.Pak.Version, r.label(), r.StatusCode, r.URL)
			}
		}
	}

	fmt.Printf("\n--- SUMMARY ---\n")
	fmt.Printf("Total: %d  |  Passed: %d  |  Failed: %d\n", len(activePaks), len(passed), len(failed))

	if len(failed) > 0 {
		os.Exit(1)
	}
}

func verify(client *http.Client, p models.Pak) Result {
	url := fmt.Sprintf("%s/releases/download/%s/%s", p.RepoURL, p.Version, p.ReleaseFilename)

	var r Result
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 5 * time.Second)
		}
		r = check(client, p, url)
		if !r.transient() {
			break
		}
	}
	return r
}

func check(client *http.Client, p models.Pak, url string) Result {
	resp, err := client.Head(url)
	if err != nil {
		return Result{Pak: p, URL: url, Error: err}
	}
	resp.Body.Close()
	return Result{Pak: p, URL: url, StatusCode: resp.StatusCode}
}
