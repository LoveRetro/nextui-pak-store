package state

import (
	"context"
	"database/sql"
	"slices"
	"strconv"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/internal"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/utils"
)

func GetInstalledPaks() (map[string]database.InstalledPak, error) {
	ctx := context.Background()
	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		if p.PakID.Valid && p.PakID.String != "" {
			installedMap[p.PakID.String] = p
		} else if p.RepoUrl.Valid && p.RepoUrl.String != "" {
			installedMap[p.RepoUrl.String] = p
		}
	}

	delete(installedMap, "Pak Store")
	delete(installedMap, models.PakStoreID)

	return installedMap, nil
}

type PakWithStatus struct {
	Pak         models.Pak
	IsInstalled bool
	HasUpdate   bool
}

func GetBrowsePaks(storefront models.Storefront, installedPaks map[string]database.InstalledPak) map[string]map[string]PakWithStatus {
	browsePaks := make(map[string]map[string]PakWithStatus)
	currentPlatform := utils.GetPlatform()
	config := internal.GetConfig()

	for _, p := range storefront.Paks {
		if p.Disabled {
			continue
		}

		if config.PlatformFilter == internal.PlatformFilterMatchDevice && !supportsCurrentPlatform(p, currentPlatform) {
			continue
		}

		installed, isInstalled := findInstalledPak(p, installedPaks)
		hasUpdate := false
		if isInstalled {
			hasUpdate = HasUpdate(installed.Version, p.Version)
		}

		pakStatus := PakWithStatus{
			Pak:         p,
			IsInstalled: isInstalled,
			HasUpdate:   hasUpdate,
		}

		for _, cat := range p.Categories {
			if _, ok := browsePaks[cat]; !ok {
				browsePaks[cat] = make(map[string]PakWithStatus)
			}
			browsePaks[cat][p.StorefrontName] = pakStatus
		}
	}

	return browsePaks
}

func findInstalledPak(pak models.Pak, installedPaks map[string]database.InstalledPak) (database.InstalledPak, bool) {
	if pak.ID != "" {
		if installed, ok := installedPaks[pak.ID]; ok {
			return installed, true
		}
	}

	if pak.RepoURL != "" {
		if installed, ok := installedPaks[pak.RepoURL]; ok {
			return installed, true
		}
	}

	return database.InstalledPak{}, false
}

func GetUpdatesAvailable(storefront models.Storefront, installedPaks map[string]database.InstalledPak) []models.Pak {
	var updates []models.Pak
	currentPlatform := utils.GetPlatform()
	config := internal.GetConfig()

	for _, p := range storefront.Paks {
		if config.PlatformFilter == internal.PlatformFilterMatchDevice && !supportsCurrentPlatform(p, currentPlatform) {
			continue
		}

		installed, isInstalled := findInstalledPak(p, installedPaks)
		if isInstalled && HasUpdate(installed.Version, p.Version) {
			updates = append(updates, p)
		}
	}

	slices.SortFunc(updates, func(a, b models.Pak) int {
		return strings.Compare(a.StorefrontName, b.StorefrontName)
	})

	return updates
}

func MigratePreID(storefront models.Storefront) error {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	installed, err := database.DBQ().ListInstalledPaksWithoutPakID(ctx)
	if err != nil {
		return err
	}

	for _, p := range installed {
		for _, sfp := range storefront.Paks {
			matched := false
			matchReason := ""

			if p.RepoUrl.Valid && p.RepoUrl.String != "" && p.RepoUrl.String == sfp.RepoURL {
				matched = true
				matchReason = "repo_url"
			}

			if !matched && p.RepoUrl.Valid && p.RepoUrl.String != "" {
				if slices.Contains(sfp.PreviousRepoURLs, p.RepoUrl.String) {
					matched = true
					matchReason = "previous_repo_url"
				}
			}

			if !matched && p.RepoUrl.String == "" {
				if p.DisplayName == sfp.StorefrontName || slices.Contains(sfp.PreviousNames, p.DisplayName) {
					matched = true
					matchReason = "display_name"
				}
			}

			if matched {
				if p.RepoUrl.Valid && p.RepoUrl.String != "" {
					err := database.DBQ().UpdateInstalledWithPakID(ctx, database.UpdateInstalledWithPakIDParams{
						PakID:          sql.NullString{String: sfp.ID, Valid: true},
						NewDisplayName: sfp.StorefrontName,
						NewName:        sfp.Name,
						NewRepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
						OldRepoUrl:     p.RepoUrl,
					})
					if err != nil {
						logger.Error("Failed to update installed pak with pak_id", "error", err)
					} else {
						logger.Info("Updated installed Pak with pak_id",
							"pak", p.DisplayName,
							"pak_id", sfp.ID,
							"match_reason", matchReason)
					}
				} else {
					err := database.DBQ().UpdateInstalledWithRepo(ctx, database.UpdateInstalledWithRepoParams{
						NewDisplayName: sfp.StorefrontName,
						NewName:        sfp.Name,
						NewRepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
						OldDisplayName: p.DisplayName,
					})
					if err != nil {
						logger.Error("Failed to update installed pak with repo URL", "error", err)
					} else {
						logger.Info("Updated installed Pak with Repo URL",
							"pak", p.DisplayName,
							"repo", sfp.RepoURL,
							"match_reason", matchReason)
					}
				}
				break
			}
		}
	}

	return nil
}

func SyncInstalledMetadataFromStorefront(storefront models.Storefront) error {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	installed, err := database.DBQ().ListInstalledPaksWithPakID(ctx)
	if err != nil {
		return err
	}

	storefrontByID := make(map[string]models.Pak)
	for _, sfp := range storefront.Paks {
		if sfp.ID != "" {
			storefrontByID[sfp.ID] = sfp
		}
	}

	for _, p := range installed {
		if !p.PakID.Valid || p.PakID.String == "" {
			continue
		}

		sfp, found := storefrontByID[p.PakID.String]
		if !found {
			continue
		}

		needsUpdate := p.DisplayName != sfp.StorefrontName ||
			p.Name != sfp.Name ||
			!p.RepoUrl.Valid ||
			p.RepoUrl.String != sfp.RepoURL

		if needsUpdate {
			err := database.DBQ().SyncInstalledByPakID(ctx, database.SyncInstalledByPakIDParams{
				DisplayName: sfp.StorefrontName,
				Name:        sfp.Name,
				RepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
				PakID:       p.PakID,
			})
			if err != nil {
				logger.Error("Failed to sync installed pak data", "error", err, "pak_id", p.PakID.String)
			} else {
				logger.Info("Synced installed pak data from storefront",
					"pak_id", p.PakID.String,
					"old_name", p.DisplayName,
					"new_name", sfp.StorefrontName)
			}
		}
	}

	return nil
}

func supportsCurrentPlatform(pak models.Pak, platform string) bool {
	if len(pak.Platforms) == 0 {
		return true
	}
	return slices.ContainsFunc(pak.Platforms, func(p string) bool {
		return strings.EqualFold(p, platform)
	})
}

func HasUpdate(installed string, latest string) bool {
	return compareVersions(installed, latest) == -1
}

func compareVersions(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int

		if i < len(partsA) {
			numA, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			numB, _ = strconv.Atoi(partsB[i])
		}

		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}

	return 0
}
