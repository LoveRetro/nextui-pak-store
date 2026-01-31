package ui

import (
	"errors"
	"slices"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
)

type ManageInstalledInput struct {
	Storefront           models.Storefront
	LastSelectedIndex    int
	LastSelectedPosition int
}

type ManageInstalledOutput struct {
	SelectedPak          models.Pak
	LastSelectedIndex    int
	LastSelectedPosition int
}

type ManageInstalledScreen struct{}

func NewManageInstalledScreen() *ManageInstalledScreen {
	return &ManageInstalledScreen{}
}

func (s *ManageInstalledScreen) Draw(input ManageInstalledInput) (ScreenResult[ManageInstalledOutput], error) {
	output := ManageInstalledOutput{
		LastSelectedIndex:    input.LastSelectedIndex,
		LastSelectedPosition: input.LastSelectedPosition,
	}

	installedPaks, err := state.GetUninstallablePaks()
	if err != nil {
		return withAction(output, ActionError), err
	}

	if len(installedPaks) == 0 {
		return back(output), nil
	}

	var menuItems []gaba.MenuItem

	for _, installed := range installedPaks {
		var pak models.Pak

		for _, p := range input.Storefront.Paks {
			// Match by pak_id first, then fall back to repo_url
			matched := false
			if installed.PakID.Valid && installed.PakID.String != "" && p.ID == installed.PakID.String {
				matched = true
			} else if installed.RepoUrl.Valid && p.RepoURL == installed.RepoUrl.String {
				matched = true
			}

			if matched {
				pak = p
				break
			}
		}

		if pak.StorefrontName != "" {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     pak.StorefrontName,
				Selected: false,
				Focused:  false,
				Metadata: pak,
			})
		}
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Manage Installed Paks", menuItems)
	options.SelectedIndex = input.LastSelectedIndex
	options.VisibleStartIndex = max(0, input.LastSelectedIndex-input.LastSelectedPosition)
	options.FooterHelpItems = BackSelectFooter()

	sel, err := gaba.List(options)
	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		return withAction(output, ActionError), err
	}

	if len(sel.Selected) == 0 {
		return back(output), nil
	}

	output.SelectedPak = sel.Items[sel.Selected[0]].Metadata.(models.Pak)
	output.LastSelectedIndex = sel.Selected[0]
	output.LastSelectedPosition = sel.VisiblePosition

	return success(output), nil
}
