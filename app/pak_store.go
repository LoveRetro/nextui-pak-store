package main

import (
	_ "embed"
	"log/slog"
	"path/filepath"
	"time"

	_ "github.com/BrandonKowalski/certifiable"
	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
	"github.com/LoveRetro/nextui-pak-store/ui"
	"github.com/LoveRetro/nextui-pak-store/utils"
	_ "modernc.org/sqlite"
)

var storefront models.Storefront
var experimentalUnlocked bool

func init() {
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
		gaba.ProcessMessageOptions{Image: "resources/splash.png", ImageWidth: 1024, ImageHeight: 768}, func() (models.Storefront, error) {
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

		gaba.ConfirmationMessage("Experimental Paks Unlocked.\n"+ui.ExperimentalWarning, []gaba.FooterHelpItem{
			{ButtonName: "A", HelpText: "Continue"},
		}, gaba.MessageOptions{})
	}

	if err != nil {
		gaba.ConfirmationMessage("Could not load the Storefront!\nMake sure you are connected to Wi-Fi.\nIf this issue persists, check the logs.", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
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
