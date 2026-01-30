package ui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/database"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/utils"
)

type PakInfoInput struct {
	Paks        []models.Pak
	Category    string
	IsUpdate    bool
	IsInstalled bool
}

type PakInfoOutput struct {
	IsUpdate     bool
	WasInstalled bool
}

type PakInfoScreen struct{}

func NewPakInfoScreen() *PakInfoScreen {
	return &PakInfoScreen{}
}

func (s *PakInfoScreen) Draw(input PakInfoInput) (ScreenResult[PakInfoOutput], error) {
	if len(input.Paks) == 1 {
		return s.drawSingle(input)
	}
	return s.drawMultiple(input)
}

func (s *PakInfoScreen) drawSingle(input PakInfoInput) (ScreenResult[PakInfoOutput], error) {
	logger := gaba.GetLogger()
	output := PakInfoOutput{IsUpdate: input.IsUpdate}

	pak := input.Paks[0]

	screenshots := make([]string, len(pak.Screenshots))

	const maxConcurrentDownloads = 4
	sem := make(chan struct{}, maxConcurrentDownloads)

	var wg sync.WaitGroup

	for i, screenshot := range pak.Screenshots {
		wg.Add(1)
		go func(index int, screenshot string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			uri := pak.RepoURL + models.RefMainStub + screenshot
			uri = strings.ReplaceAll(uri, models.GitHubRoot, models.RawGHUC)

			downloadedScreenshot, err := utils.DownloadTempFile(uri)
			if err == nil {
				screenshots[index] = downloadedScreenshot
			} else {
				logger.Error("Failed to download screenshot",
					"error", err,
					"uri", uri,
					"attempt", 1)

				downloadedScreenshot, err = utils.DownloadTempFile(uri)
				if err == nil {
					screenshots[index] = downloadedScreenshot
				} else {
					logger.Error("Failed to download screenshot after retry",
						"error", err,
						"uri", uri)
				}
			}
		}(i, screenshot)
	}

	wg.Wait()

	filteredScreenshots := make([]string, 0, len(screenshots))
	for _, s := range screenshots {
		if s != "" {
			filteredScreenshots = append(filteredScreenshots, s)
		}
	}
	screenshots = filteredScreenshots

	var sections []gaba.Section

	if _, ok := pak.Changelog[pak.Version]; ok && input.IsUpdate {
		sections = append(sections,
			gaba.NewDescriptionSection(
				fmt.Sprintf("What's new in %s?", pak.Version),
				pak.Changelog[pak.Version],
			))
	}

	if pak.Description != "" {
		sections = append(sections, gaba.NewDescriptionSection(
			"Description",
			pak.Description,
		))
	}

	if len(screenshots) > 0 {
		sections = append(sections, gaba.NewSlideshowSection(
			"Screenshots",
			screenshots,
			int32(float64(gaba.GetWindow().GetWidth())/1.2),
			int32(float64(gaba.GetWindow().GetHeight())/1.2),
		))
	}

	sections = append(sections, gaba.NewInfoSection(
		"Pak Info",
		[]gaba.MetadataItem{
			{Label: "Author", Value: pak.Author},
			{Label: "Version", Value: pak.Version},
		},
	))

	var changelog []string

	var versions []string
	for k := range pak.Changelog {
		versions = append(versions, k)
	}

	slices.SortFunc(versions, func(a, b string) int {
		return strings.Compare(b, a)
	})

	for _, v := range versions {
		changelog = append(changelog, fmt.Sprintf("%s: %s", v, pak.Changelog[v]))
	}

	if len(changelog) > 0 {
		sections = append(sections, gaba.NewDescriptionSection(
			"Changelog",
			strings.Join(changelog, "\n\n"),
		))
	}

	qrcode, err := utils.CreateTempQRCode(pak.RepoURL, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"Pak Repository",
			qrcode,
			int32(256),
			int32(256),
			constants.TextAlignCenter,
		))

	} else {
		logger.Error("Unable to generate QR code", "error", err)
	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.AllowAction = true

	var footerItems []gaba.FooterHelpItem
	if input.IsUpdate {
		footerItems = []gaba.FooterHelpItem{
			FooterBack(),
			{ButtonName: "A", HelpText: "Update"},
		}
	} else if input.IsInstalled {
		footerItems = []gaba.FooterHelpItem{
			FooterBack(),
			{ButtonName: "A", HelpText: "Uninstall"},
		}
	} else {
		footerItems = []gaba.FooterHelpItem{
			FooterBack(),
			{ButtonName: "A", HelpText: "Install"},
		}
	}

	_, err = gaba.DetailScreen(pak.StorefrontName, options, footerItems)
	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		logger.Error("Unable to display pak info screen", "error", err)
		return withAction(output, ActionError), err
	}

	if input.IsInstalled && !input.IsUpdate {
		_, err = gaba.ConfirmationMessage(fmt.Sprintf("Are you sure that you want to uninstall\n %s?", pak.Name),
			[]gaba.FooterHelpItem{
				{ButtonName: "B", HelpText: "Nevermind"},
				{ButtonName: "X", HelpText: "Yes"},
			}, gaba.MessageOptions{
				ConfirmButton: constants.VirtualButtonX,
			})

		if err != nil {
			if errors.Is(err, gaba.ErrCancelled) {
				return withAction(output, ActionCancelled), nil
			}
			return withAction(output, ActionError), err
		}

		_, err = gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Uninstalling", pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			pakLocation := ""

			switch pak.PakType {
			case models.PakTypes.TOOL:
				pakLocation = filepath.Join(utils.GetToolRoot(), pak.Name+".pak")
			case models.PakTypes.EMU:
				pakLocation = filepath.Join(utils.GetEmulatorRoot(), pak.Name+".pak")
			}

			err = os.RemoveAll(pakLocation)

			time.Sleep(1750 * time.Millisecond)

			return nil, err
		})

		if err != nil {
			gaba.ProcessMessage(fmt.Sprintf("Unable to uninstall %s", pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, nil
			})
			logger.Error("Unable to remove pak", "error", err)
		}

		ctx := context.Background()
		err = database.DBQ().Uninstall(ctx, sql.NullString{String: pak.ID, Valid: true})
		if err != nil {
			logger.Error("Failed to uninstall pak from database", "error", err)
		}

		output.WasInstalled = true
		return withAction(output, ActionUninstalled), nil
	}

	tmp, completed, err := utils.DownloadPakArchive(pak)
	if err != nil {

		if err.Error() == "download cancelled by user" {
			return withAction(output, ActionCancelled), nil
		}

		logger.Error("Unable to download pak archive", "error", err)
		return withAction(output, ActionError), nil
	} else if !completed {
		return withAction(output, ActionCancelled), nil
	}

	err = utils.UnzipPakArchive(pak, tmp)
	if err != nil {
		logger.Error("Unable to extract pak archive", "error", err)
		gaba.ProcessMessage(fmt.Sprintf("Failed to extract %s", pak.StorefrontName),
			gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
				time.Sleep(2 * time.Second)
				return nil, nil
			})
		return withAction(output, ActionError), nil
	}

	if !input.IsUpdate {
		info := database.InstallParams{
			DisplayName:  pak.StorefrontName,
			Name:         pak.Name,
			PakID:        sql.NullString{String: pak.ID, Valid: true},
			RepoUrl:      sql.NullString{String: pak.RepoURL, Valid: true},
			Version:      pak.Version,
			Type:         models.PakTypeMap[pak.PakType],
			CanUninstall: int64(1),
		}
		database.DBQ().Install(context.Background(), info)
	} else {
		update := database.UpdateVersionParams{
			Version: pak.Version,
			RepoUrl: sql.NullString{String: pak.RepoURL, Valid: true},
			PakID:   sql.NullString{String: pak.ID, Valid: true},
		}
		database.DBQ().UpdateVersion(context.Background(), update)
	}

	action := "Installed"
	if input.IsUpdate {
		action = "Updated"
	}

	if pak.Name == "Pak Store" {
		return withAction(output, ActionPakStoreUpdated), nil
	}

	gaba.ProcessMessage(fmt.Sprintf("%s %s!", pak.StorefrontName, action), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(1250 * time.Millisecond)
		return nil, nil
	})

	return withAction(output, ActionInstallSuccess), nil
}

func (s *PakInfoScreen) drawMultiple(input PakInfoInput) (ScreenResult[PakInfoOutput], error) {
	logger := gaba.GetLogger()
	output := PakInfoOutput{IsUpdate: input.IsUpdate}

	if len(input.Paks) == 0 {
		return back(output), nil
	}

	var sections []gaba.Section

	pakNames := make([]string, len(input.Paks))
	for i, pak := range input.Paks {
		pakNames[i] = pak.StorefrontName
	}

	overviewText := fmt.Sprintf("The following %d paks will be updated!",
		len(input.Paks))

	sections = append(sections, gaba.NewDescriptionSection(
		"Update Overview",
		overviewText,
	))

	for _, pak := range input.Paks {
		info := []gaba.MetadataItem{
			{Label: "Author", Value: pak.Author},
			{Label: "Current Version", Value: pak.Version},
		}

		if changelog, ok := pak.Changelog[pak.Version]; ok {
			info = append(info, gaba.MetadataItem{Label: "Changelog", Value: changelog})
		}

		sections = append(sections, gaba.NewInfoSection(
			pak.StorefrontName,
			info,
		))

	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ConfirmButton = constants.VirtualButtonX

	footerItems := []gaba.FooterHelpItem{
		FooterCancel(),
		{ButtonName: "X", HelpText: "Update All"},
	}

	title := fmt.Sprintf("Update %d Paks", len(input.Paks))

	var err error
	_, err = gaba.DetailScreen(title, options, footerItems)
	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		logger.Error("Unable to display multi-pak info screen", "error", err)
		return withAction(output, ActionError), err
	}

	for _, pak := range input.Paks {
		tmp, completed, err := utils.DownloadPakArchive(pak)
		if err != nil {
			if err.Error() == "download cancelled by user" {
				return withAction(output, ActionPartialUpdate), nil
			}
			logger.Error("Failed to download pak",
				"error", err,
				"pak", pak.StorefrontName)
			gaba.ProcessMessage(fmt.Sprintf("Failed to download %s", pak.StorefrontName),
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			continue
		} else if !completed {
			return withAction(output, ActionPartialUpdate), nil
		}

		err = utils.UnzipPakArchive(pak, tmp)
		if err != nil {
			logger.Error("Failed to extract pak",
				"error", err,
				"pak", pak.StorefrontName)
			gaba.ProcessMessage(fmt.Sprintf("Failed to extract %s", pak.StorefrontName),
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			continue
		}

		update := database.UpdateVersionParams{
			Version: pak.Version,
			RepoUrl: sql.NullString{String: pak.RepoURL, Valid: true},
			PakID:   sql.NullString{String: pak.ID, Valid: true},
		}
		err = database.DBQ().UpdateVersion(context.Background(), update)
		if err != nil {
			logger.Error("Failed to update pak in database",
				"error", err,
				"pak", pak.Name)
		}

		if pak.Name == "Pak Store" {
			gaba.ProcessMessage("Pak Store Updated! Restarting...",
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			return withAction(output, ActionPakStoreUpdated), nil
		}
	}

	gaba.ProcessMessage("All paks updated successfully!",
		gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
			time.Sleep(2 * time.Second)
			return nil, nil
		})

	return success(output), nil
}
