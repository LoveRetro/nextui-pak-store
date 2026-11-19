package state

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/models"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"alpha patch update", "v1.0.22-alpha", "v1.0.23-alpha", -1},
		{"alpha to stable", "v1.0.23-alpha", "v1.0.23", -1},
		{"numeric prerelease", "1.0.0-alpha.2", "1.0.0-alpha.10", -1},
		{"numeric before text", "1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"prerelease text order", "1.0.0-alpha", "1.0.0-beta", -1},
		{"prerelease identifier count", "1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"build metadata ignored", "1.0.0-alpha+first", "v1.0.0-alpha+second", 0},
		{"optional prefix and zero revision", "1.0.0", "v1.0.0.0", 0},
		{"fourth component", "v5.4.1.0", "v5.4.1.1", -1},
		{"fourth before prerelease", "v1.0.0.1-alpha", "v1.0.0", 1},
		{"fourth with prerelease", "v1.0.0.1-alpha", "v1.0.0.1", -1},
		{"fourth leading zeros", "1.0.0.001", "1.0.0.1", 0},
		{"large core numbers", "1.0.99999999999999999999-alpha", "1.0.100000000000000000000-alpha", -1},
		{"large prerelease numbers", "1.0.0-alpha.99999999999999999999", "1.0.0-alpha.100000000000000000000", -1},
		{"large fourth numbers", "1.0.0.99999999999999999999", "1.0.0.100000000000000000000", -1},
		{"legacy leading zeros", "v1.02.3", "v1.2.4", -1},
		{"legacy malformed component", "v1.0.bad", "v1.0.1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareVersions(tt.a, tt.b); got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
			if got := compareVersions(tt.b, tt.a); got != -tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.b, tt.a, got, -tt.want)
			}
			if got := HasUpdate(tt.a, tt.b); got != (tt.want < 0) {
				t.Errorf("HasUpdate(%q, %q) = %t", tt.a, tt.b, got)
			}
			if got := HasUpdate(tt.b, tt.a); got != (tt.want > 0) {
				t.Errorf("HasUpdate(%q, %q) = %t", tt.b, tt.a, got)
			}
		})
	}
}

func TestInstalledPakShowsAlphaUpdate(t *testing.T) {
	t.Setenv("ENVIRONMENT", "DEV")
	t.Setenv("PLATFORM", "tg5040")
	t.Chdir(t.TempDir())
	if err := os.WriteFile("pak.json", []byte(`{"name":"Pak Store","version":"v3.2.0"}`), 0644); err != nil {
		t.Fatal(err)
	}
	database.Init()
	t.Cleanup(database.CloseDB)

	pak := models.Pak{
		ID:             "lmwlIooQTW",
		StorefrontName: "ChooChooTracker",
		Name:           "ChooChooTracker",
		Version:        "v1.0.25-alpha",
		RepoURL:        "https://github.com/paiheulevrai/Choochootracker",
		Platforms:      []string{"tg5040"},
		Categories:     []string{"Media"},
	}
	storefront := models.Storefront{Paks: []models.Pak{pak}}
	pakID := sql.NullString{String: pak.ID, Valid: true}
	repoURL := sql.NullString{String: pak.RepoURL, Valid: true}
	ctx := context.Background()
	if err := database.DBQ().Install(ctx, database.InstallParams{
		DisplayName: pak.StorefrontName, Name: pak.Name, PakID: pakID,
		RepoUrl: repoURL, Version: "v1.0.24-alpha", Type: "TOOL", CanUninstall: 1,
	}); err != nil {
		t.Fatal(err)
	}

	installed, err := GetInstalledPaks()
	if err != nil {
		t.Fatal(err)
	}
	status := GetBrowsePaks(storefront, installed, false)["Media"][pak.StorefrontName]
	if !status.IsInstalled || !status.HasUpdate {
		t.Errorf("installed CCT should show an alpha update: %+v", status)
	}
	if updates := GetUpdatesAvailable(storefront, false); len(updates) != 1 || updates[0].ID != pak.ID {
		t.Errorf("alpha update should appear in updates: %+v", updates)
	}

	if err := database.DBQ().UpdateVersion(ctx, database.UpdateVersionParams{
		Version: pak.Version, RepoUrl: repoURL, PakID: pakID,
	}); err != nil {
		t.Fatal(err)
	}
	installed, err = GetInstalledPaks()
	if err != nil {
		t.Fatal(err)
	}
	status = GetBrowsePaks(storefront, installed, false)["Media"][pak.StorefrontName]
	if !status.IsInstalled || status.HasUpdate {
		t.Fatalf("current CCT should not show an update: %+v", status)
	}
	if updates := GetUpdatesAvailable(storefront, false); len(updates) != 0 {
		t.Fatalf("current CCT should not appear in updates: %+v", updates)
	}
}
