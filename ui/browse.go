package ui

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/state"
)

type BrowseInput struct {
	Storefront           models.Storefront
	LastSelectedIndex    int
	LastSelectedPosition int
}

type BrowseOutput struct {
	SelectedCategory     string
	LastSelectedIndex    int
	LastSelectedPosition int
}

type BrowseScreen struct{}

func NewBrowseScreen() *BrowseScreen {
	return &BrowseScreen{}
}

func (s *BrowseScreen) Draw(input BrowseInput) (ScreenResult[BrowseOutput], error) {
	output := BrowseOutput{
		LastSelectedIndex:    input.LastSelectedIndex,
		LastSelectedPosition: input.LastSelectedPosition,
	}

	// Compute data on demand
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withAction(output, ActionError), err
	}

	browsePaks := state.GetBrowsePaks(input.Storefront, installedPaks)

	var menuItems []gaba.MenuItem

	for cat := range browsePaks {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     cat + " (" + strconv.Itoa(len(browsePaks[cat])) + ")",
			Selected: false,
			Focused:  false,
			Metadata: cat,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Browse Paks", menuItems)
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

	output.SelectedCategory = sel.Items[sel.Selected[0]].Metadata.(string)
	output.LastSelectedIndex = sel.Selected[0]
	output.LastSelectedPosition = sel.VisiblePosition

	return success(output), nil
}
