package ui

import (
	"errors"
	"fmt"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
)

type MainMenuInput struct {
	Storefront models.Storefront
}

type MainMenuOutput struct {
	Selection string
}

type MainMenuScreen struct{}

func NewMainMenuScreen() *MainMenuScreen {
	return &MainMenuScreen{}
}

func (s *MainMenuScreen) Draw(input MainMenuInput) (ScreenResult[MainMenuOutput], error) {
	output := MainMenuOutput{}

	// Compute data on demand
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withAction(output, ActionError), err
	}

	browsePaks := state.GetBrowsePaks(input.Storefront, installedPaks)
	updatesAvailable := state.GetUpdatesAvailable(input.Storefront)

	title := "Pak Store"

	var menuItems []gaba.MenuItem

	if len(updatesAvailable) > 0 {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     fmt.Sprintf("Available Updates (%d)", len(updatesAvailable)),
			Selected: false,
			Focused:  false,
			Metadata: "Available Updates",
		})
	}

	if len(browsePaks) > 0 {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     "Browse",
			Selected: false,
			Focused:  false,
			Metadata: "Browse",
		})
	}

	if len(installedPaks) > 0 {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     "Manage Installed",
			Selected: false,
			Focused:  false,
			Metadata: "Manage Installed",
		})
	}

	options := gaba.DefaultListOptions(title, menuItems)
	options.FooterHelpItems = []gaba.FooterHelpItem{
		FooterQuit(),
		{ButtonName: "X", HelpText: "Settings"},
		FooterSelect(),
	}
	options.ActionButton = constants.VirtualButtonX

	options.EmptyMessage = "No Paks Available"

	sel, err := gaba.List(options)
	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return withAction(output, ActionQuit), nil
		}
		return withAction(output, ActionError), err
	}

	// Handle X button for Settings
	if sel.Action == gaba.ListActionTriggered {
		return withAction(output, ActionSettings), nil
	}

	if len(sel.Selected) == 0 {
		return withAction(output, ActionQuit), nil
	}

	output.Selection = sel.Items[sel.Selected[0]].Metadata.(string)

	switch output.Selection {
	case "Browse":
		return withAction(output, ActionBrowse), nil
	case "Available Updates":
		return withAction(output, ActionUpdates), nil
	case "Manage Installed":
		return withAction(output, ActionManageInstalled), nil
	}

	return success(output), nil
}
