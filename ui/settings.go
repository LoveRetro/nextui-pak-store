package ui

import (
	"errors"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/internal"
)

const ExperimentalWarning = "Use at your own risk!\nMake sure you have backups!"

type SettingsInput struct {
	Config           *internal.Config
	ExperimentalMode bool
}

type SettingsOutput struct {
	Config           *internal.Config
	ExperimentalMode bool
}

type SettingsScreen struct{}

func NewSettingsScreen() *SettingsScreen {
	return &SettingsScreen{}
}

func (s *SettingsScreen) Draw(input SettingsInput) (ScreenResult[SettingsOutput], error) {
	config := input.Config
	output := SettingsOutput{Config: config, ExperimentalMode: input.ExperimentalMode}

	wasDiscoverEnabled := config.ShouldDiscoverExistingInstalls()

	items := s.buildMenuItems(config, input.ExperimentalMode)

	result, err := gaba.OptionsList(
		"Settings",
		gaba.OptionListSettings{
			FooterHelpItems: OptionsListFooter(),
			UseSmallTitle:   true,
		},
		items,
	)

	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		gaba.GetLogger().Error("Settings error", "error", err)
		return withAction(output, ActionError), err
	}

	// Check if Info was clicked
	if result.Action == gaba.ListActionSelected {
		selectedText := items[result.Selected].Item.Text
		if selectedText == "Info" {
			return withAction(output, ActionInfo), nil
		}
	}

	s.applySettings(config, result.Items)

	// Experimental mode only lasts for this session, so it is not saved to the config
	output.ExperimentalMode = s.resolveExperimentalMode(input.ExperimentalMode, result.Items)

	err = internal.SaveConfig(config)
	if err != nil {
		gaba.GetLogger().Error("Error saving settings", "error", err)
		return withAction(output, ActionError), err
	}

	// If discover was just turned on, trigger a scan
	if !wasDiscoverEnabled && config.ShouldDiscoverExistingInstalls() {
		return withAction(output, ActionDiscoverExistingInstalls), nil
	}

	return withAction(output, ActionSettingsSaved), nil
}

func (s *SettingsScreen) resolveExperimentalMode(wasEnabled bool, items []gaba.ItemWithOptions) bool {
	enabled := wasEnabled
	for _, item := range items {
		if item.Item.Text != "Experimental Paks" {
			continue
		}
		if val, ok := item.Options[item.SelectedOption].Value.(bool); ok {
			enabled = val
		}
	}

	if !enabled || wasEnabled {
		return enabled
	}

	_, err := gaba.ConfirmationMessage("Enable Experimental Paks?\n"+ExperimentalWarning, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "A", HelpText: "Enable"},
	}, gaba.MessageOptions{})
	if err != nil {
		if !errors.Is(err, gaba.ErrCancelled) {
			gaba.GetLogger().Error("Experimental Paks confirmation error", "error", err)
		}
		return false
	}

	return true
}

func (s *SettingsScreen) buildMenuItems(config *internal.Config, experimentalMode bool) []gaba.ItemWithOptions {
	return []gaba.ItemWithOptions{
		{
			Item: gaba.MenuItem{Text: "Platform Filter"},
			Options: []gaba.Option{
				{DisplayName: "Match Device", Value: internal.PlatformFilterMatchDevice},
				{DisplayName: "All", Value: internal.PlatformFilterAll},
			},
			SelectedOption: platformFilterToIndex(config.PlatformFilter),
		},
		{
			Item: gaba.MenuItem{Text: "Debug Level"},
			Options: []gaba.Option{
				{DisplayName: "Error", Value: internal.DebugLevelError},
				{DisplayName: "Info", Value: internal.DebugLevelInfo},
				{DisplayName: "Debug", Value: internal.DebugLevelDebug},
			},
			SelectedOption: debugLevelToIndex(config.DebugLevel),
		},
		{
			Item: gaba.MenuItem{Text: "Discover Existing Installs"},
			Options: []gaba.Option{
				{DisplayName: "On", Value: true},
				{DisplayName: "Off", Value: false},
			},
			SelectedOption: discoverExistingInstallsToIndex(config.ShouldDiscoverExistingInstalls()),
		},
		{
			Item: gaba.MenuItem{Text: "Experimental Paks"},
			Options: []gaba.Option{
				{DisplayName: "On", Value: true},
				{DisplayName: "Off", Value: false},
			},
			SelectedOption: experimentalModeToIndex(experimentalMode),
		},
		{
			Item:    gaba.MenuItem{Text: "Info"},
			Options: []gaba.Option{{Type: gaba.OptionTypeClickable}},
		},
	}
}

func (s *SettingsScreen) applySettings(config *internal.Config, items []gaba.ItemWithOptions) {
	for _, item := range items {
		switch item.Item.Text {
		case "Platform Filter":
			if val, ok := item.Options[item.SelectedOption].Value.(internal.PlatformFilterMode); ok {
				config.PlatformFilter = val
			}
		case "Debug Level":
			if val, ok := item.Options[item.SelectedOption].Value.(internal.DebugLevel); ok {
				config.DebugLevel = val
			}
		case "Discover Existing Installs":
			if val, ok := item.Options[item.SelectedOption].Value.(bool); ok {
				config.DiscoverExistingInstalls = &val
			}
		}
	}
}

func platformFilterToIndex(mode internal.PlatformFilterMode) int {
	switch mode {
	case internal.PlatformFilterMatchDevice:
		return 0
	case internal.PlatformFilterAll:
		return 1
	default:
		return 0
	}
}

func debugLevelToIndex(level internal.DebugLevel) int {
	switch level {
	case internal.DebugLevelError:
		return 0
	case internal.DebugLevelInfo:
		return 1
	case internal.DebugLevelDebug:
		return 2
	default:
		return 0
	}
}

func discoverExistingInstallsToIndex(enabled bool) int {
	if enabled {
		return 0
	}
	return 1
}

func experimentalModeToIndex(enabled bool) int {
	if enabled {
		return 0
	}
	return 1
}
