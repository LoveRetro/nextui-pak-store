package ui

import (
	"errors"
	"slices"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/internal/i18n"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
)

type UpdatesInput struct {
	Storefront           models.Storefront
	ExperimentalMode     bool
	LastSelectedIndex    int
	LastSelectedPosition int
}

type UpdatesOutput struct {
	SelectedPaks         []models.Pak
	LastSelectedIndex    int
	LastSelectedPosition int
}

type UpdatesScreen struct{}

func NewUpdatesScreen() *UpdatesScreen {
	return &UpdatesScreen{}
}

func (s *UpdatesScreen) Draw(input UpdatesInput) (ScreenResult[UpdatesOutput], error) {
	output := UpdatesOutput{
		LastSelectedIndex:    input.LastSelectedIndex,
		LastSelectedPosition: input.LastSelectedPosition,
	}

	updatesAvailable := state.GetUpdatesAvailable(input.Storefront, input.ExperimentalMode)

	if len(updatesAvailable) == 0 {
		return back(output), nil
	}

	var menuItems []gaba.MenuItem

	for _, pak := range updatesAvailable {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: []models.Pak{pak},
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	if len(menuItems) > 1 {
		menuItems = append([]gaba.MenuItem{{
			Text:     i18n.T("ps.updates.update_all"),
			Selected: false,
			Focused:  false,
			Metadata: updatesAvailable,
		}}, menuItems...)
	}

	options := gaba.DefaultListOptions(i18n.T("ps.title.updates"), menuItems)
	options.SelectedIndex = input.LastSelectedIndex
	options.VisibleStartIndex = max(0, input.LastSelectedIndex-input.LastSelectedPosition)
	options.FooterHelpItems = BackViewFooter()

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

	output.SelectedPaks = sel.Items[sel.Selected[0]].Metadata.([]models.Pak)
	output.LastSelectedIndex = sel.Selected[0]
	output.LastSelectedPosition = sel.VisiblePosition

	return success(output), nil
}
