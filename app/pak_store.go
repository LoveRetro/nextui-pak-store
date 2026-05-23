package main

import (
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/BrandonKowalski/certifiable"
	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/internal/i18n"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
	"github.com/LoveRetro/nextui-pak-store/utils"
	_ "modernc.org/sqlite"
)

// splashForLocale returns the locale-specific splash image path when it
// exists on disk, falling back to the English splash. The active locale
// comes from NextUI's minuisettings (via i18n.Active()).
func splashForLocale() string {
	code := i18n.Active()
	if code != "" && code != "en" {
		candidate := "resources/splash_" + code + ".png"
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "resources/splash.png"
}

var storefront models.Storefront
var experimentalUnlocked bool

func init() {
	i18n.Init()

	logPath := filepath.Join(utils.GetLogsDir(), "pak_store.log")
	gaba.Init(gaba.Options{
		WindowTitle:    "Pak Store",
		ShowBackground: true,
		LogPath:        logPath,
		IsNextUI:       true,
	})

	gaba.SetLogLevel(slog.LevelDebug)

	gaba.RegisterChord("experimental", []constants.VirtualButton{
		constants.VirtualButtonL1,
		constants.VirtualButtonR1,
		constants.VirtualButtonStart,
	}, gaba.ChordOptions{
		OnTrigger: func() {
			experimentalUnlocked = true
		},
	})

	sf, err := gaba.ProcessMessage("",
		gaba.ProcessMessageOptions{Image: splashForLocale(), ImageWidth: 1024, ImageHeight: 768}, func() (models.Storefront, error) {
			time.Sleep(3 * time.Second)
			return utils.FetchStorefront()
		})

	if experimentalUnlocked {
		gaba.ProcessMessage("", gaba.ProcessMessageOptions{
			Image:       "resources/jankstore.png",
			ImageWidth:  1024,
			ImageHeight: 768,
		}, func() (any, error) {
			time.Sleep(2 * time.Second)
			return nil, nil
		})

		gaba.ConfirmationMessage(i18n.T("ps.experimental_unlocked"), []gaba.FooterHelpItem{
			{ButtonName: "A", HelpText: i18n.T("ps.btn.continue")},
		}, gaba.MessageOptions{})
	}

	if err != nil {
		gaba.ConfirmationMessage(i18n.T("ps.storefront_error"), []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: i18n.T("ps.btn.quit")},
		}, gaba.MessageOptions{})
		defer gaba.Close()
		utils.LogStandardFatal("Could not load Storefront!", err)
	}

	database.Init()

	if err := state.MigratePreID(sf); err != nil {
		gaba.GetLogger().Error("Failed to migrate installed paks to use Pak ID", "error", err)
	}

	if err := state.SyncInstalledMetadataFromStorefront(sf); err != nil {
		gaba.GetLogger().Error("Failed to sync installed metadata with storefront", "error", err)
	}

	if err := state.DiscoverExistingInstalls(sf); err != nil {
		gaba.GetLogger().Error("Failed to discover existing pak installs", "error", err)
	}

	storefront = sf
}

func cleanup() {
	database.CloseDB()
	gaba.Close()
}

func main() {
	defer cleanup()

	logger := gaba.GetLogger()

	logger.Info("Starting Pak Store")

	if err := runApp(storefront); err != nil {
		logger.Error("Router error", "error", err)
	}
}
