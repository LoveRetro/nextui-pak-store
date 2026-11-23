package state

import (
	"cmp"
	"context"
	"database/sql"
	"os"
	"slices"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/internal"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/utils"
)

func GetInstalledPaks() (map[string]database.InstalledPak, error) {
	return getInstalledPaksFiltered(false)
}

func GetUninstallablePaks() (map[string]database.InstalledPak, error) {
	return getInstalledPaksFiltered(true)
}

func getInstalledPaksFiltered(excludeNonUninstallable bool) (map[string]database.InstalledPak, error) {
	ctx := context.Background()
	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		if excludeNonUninstallable && p.CanUninstall == 0 {
			continue
		}
		if p.PakID.Valid && p.PakID.String != "" {
			installedMap[p.PakID.String] = p
		} else if p.RepoUrl.Valid && p.RepoUrl.String != "" {
			installedMap[p.RepoUrl.String] = p
		}
	}

	return installedMap, nil
}

type PakWithStatus struct {
	Pak         models.Pak
	IsInstalled bool
	HasUpdate   bool
}

func GetBrowsePaks(storefront models.Storefront, installedPaks map[string]database.InstalledPak, experimentalMode bool) map[string]map[string]PakWithStatus {
	browsePaks := make(map[string]map[string]PakWithStatus)
	currentPlatform := utils.GetPlatform()
	config := internal.GetConfig()

	for _, p := range storefront.Paks {
		if p.Disabled {
			continue
		}

		if p.Experimental && !experimentalMode {
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

		categories := p.Categories
		if p.Experimental {
			categories = []string{"Experimental"}
		}

		for _, cat := range categories {
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

func GetUpdatesAvailable(storefront models.Storefront, experimentalMode bool) []models.Pak {
	var updates []models.Pak
	currentPlatform := utils.GetPlatform()
	config := internal.GetConfig()

	installedPaks, err := GetInstalledPaks()
	if err != nil {
		return updates
	}

	for _, p := range storefront.Paks {
		if p.Experimental && !experimentalMode {
			continue
		}

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

func DiscoverExistingInstalls(sf models.Storefront) error {
	config := internal.GetConfig()
	if !config.ShouldDiscoverExistingInstalls() {
		return nil
	}

	logger := gabagool.GetLogger()
	ctx := context.Background()

	installedPaks, err := GetInstalledPaks()
	if err != nil {
		return err
	}

	for _, pak := range sf.Paks {
		if _, exists := installedPaks[pak.ID]; exists {
			continue
		}

		var pakPath string
		if pak.PakType == models.PakTypes.TOOL {
			pakPath = utils.GetToolRoot() + "/" + pak.Name + ".pak"
		} else if pak.PakType == models.PakTypes.EMU {
			pakPath = utils.GetEmulatorRoot() + "/" + pak.Name + ".pak"
		} else {
			continue
		}

		if _, err := os.Stat(pakPath); err != nil {
			continue
		}

		version := "discovered"
		pakJsonPath := pakPath + "/pak.json"
		var diskPak models.Pak
		if err := utils.ParseJSONFile(pakJsonPath, &diskPak); err == nil && diskPak.Version != "" {
			version = diskPak.Version
		}

		err := database.DBQ().Install(ctx, database.InstallParams{
			DisplayName:  pak.StorefrontName,
			Name:         pak.Name,
			PakID:        sql.NullString{String: pak.ID, Valid: true},
			RepoUrl:      sql.NullString{String: pak.RepoURL, Valid: true},
			Version:      version,
			Type:         models.PakTypeMap[pak.PakType],
			CanUninstall: 1,
		})
		if err != nil {
			logger.Error("Failed to register discovered pak", "error", err, "pak", pak.StorefrontName)
		} else {
			logger.Info("Discovered existing pak install", "pak", pak.StorefrontName, "version", version)
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
	coreA, preA := parseVersion(a)
	coreB, preB := parseVersion(b)

	for i := range max(len(coreA), len(coreB)) {
		var partA, partB string
		if i < len(coreA) {
			partA = coreA[i]
		}
		if i < len(coreB) {
			partB = coreB[i]
		}
		if result := compareNumeric(partA, partB); result != 0 {
			return result
		}
	}

	switch {
	case len(preA) == 0 && len(preB) == 0:
		return 0
	case len(preA) == 0:
		return 1
	case len(preB) == 0:
		return -1
	}

	for i := range min(len(preA), len(preB)) {
		if result := comparePrereleaseIdentifier(preA[i], preB[i]); result != 0 {
			return result
		}
	}
	return cmp.Compare(len(preA), len(preB))
}

func parseVersion(version string) ([]string, []string) {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	version, _, _ = strings.Cut(version, "+")
	coreText, preText, _ := strings.Cut(version, "-")

	core := strings.Split(coreText, ".")
	for i, part := range core {
		core[i] = leadingDigits(part)
	}

	var pre []string
	if preText != "" {
		pre = strings.Split(preText, ".")
	}
	return core, pre
}

func leadingDigits(s string) string {
	end := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
	if end < 0 {
		return s
	}
	return s[:end]
}

func isNumeric(s string) bool {
	return s != "" && leadingDigits(s) == s
}

func compareNumeric(a, b string) int {
	a = strings.TrimLeft(a, "0")
	b = strings.TrimLeft(b, "0")
	if result := cmp.Compare(len(a), len(b)); result != 0 {
		return result
	}
	return strings.Compare(a, b)
}

func comparePrereleaseIdentifier(a, b string) int {
	numericA, numericB := isNumeric(a), isNumeric(b)
	switch {
	case numericA && numericB:
		return compareNumeric(a, b)
	case numericA:
		return -1
	case numericB:
		return 1
	}
	return strings.Compare(a, b)
}
