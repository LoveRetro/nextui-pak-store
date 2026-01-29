package ui

import (
	"errors"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/internal"
)

type SettingsInput struct {
	Config *internal.Config
}

type SettingsOutput struct {
	Config *internal.Config
}

type SettingsScreen struct{}

func NewSettingsScreen() *SettingsScreen {
	return &SettingsScreen{}
}

func (s *SettingsScreen) Draw(input SettingsInput) (ScreenResult[SettingsOutput], error) {
	config := input.Config
	output := SettingsOutput{Config: config}

	items := s.buildMenuItems(config)

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

	err = internal.SaveConfig(config)
	if err != nil {
		gaba.GetLogger().Error("Error saving settings", "error", err)
		return withAction(output, ActionError), err
	}

	return withAction(output, ActionSettingsSaved), nil
}

func (s *SettingsScreen) buildMenuItems(config *internal.Config) []gaba.ItemWithOptions {
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
				{DisplayName: "Off", Value: internal.DebugLevelOff},
				{DisplayName: "Error", Value: internal.DebugLevelError},
				{DisplayName: "Warn", Value: internal.DebugLevelWarn},
				{DisplayName: "Info", Value: internal.DebugLevelInfo},
				{DisplayName: "Debug", Value: internal.DebugLevelDebug},
			},
			SelectedOption: debugLevelToIndex(config.DebugLevel),
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
	case internal.DebugLevelOff:
		return 0
	case internal.DebugLevelError:
		return 1
	case internal.DebugLevelWarn:
		return 2
	case internal.DebugLevelInfo:
		return 3
	case internal.DebugLevelDebug:
		return 4
	default:
		return 0
	}
}
