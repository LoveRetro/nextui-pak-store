package ui

import (
	"errors"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/internal"
	"github.com/LoveRetro/nextui-pak-store/internal/i18n"
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

	wasDiscoverEnabled := config.ShouldDiscoverExistingInstalls()

	items := s.buildMenuItems(config)

	result, err := gaba.OptionsList(
		i18n.T("ps.title.settings"),
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
		if selectedText == i18n.T("ps.settings.info") {
			return withAction(output, ActionInfo), nil
		}
	}

	s.applySettings(config, result.Items)

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

func (s *SettingsScreen) buildMenuItems(config *internal.Config) []gaba.ItemWithOptions {
	return []gaba.ItemWithOptions{
		{
			Item: gaba.MenuItem{Text: i18n.T("ps.settings.platform_filter")},
			Options: []gaba.Option{
				{DisplayName: i18n.T("ps.settings.platform_filter.match"), Value: internal.PlatformFilterMatchDevice},
				{DisplayName: i18n.T("ps.settings.platform_filter.all"), Value: internal.PlatformFilterAll},
			},
			SelectedOption: platformFilterToIndex(config.PlatformFilter),
		},
		{
			Item: gaba.MenuItem{Text: i18n.T("ps.settings.debug_level")},
			Options: []gaba.Option{
				{DisplayName: i18n.T("ps.settings.debug_level.error"), Value: internal.DebugLevelError},
				{DisplayName: i18n.T("ps.settings.debug_level.info"), Value: internal.DebugLevelInfo},
				{DisplayName: i18n.T("ps.settings.debug_level.debug"), Value: internal.DebugLevelDebug},
			},
			SelectedOption: debugLevelToIndex(config.DebugLevel),
		},
		{
			Item: gaba.MenuItem{Text: i18n.T("ps.settings.discover")},
			Options: []gaba.Option{
				{DisplayName: i18n.T("ps.toggle.on"), Value: true},
				{DisplayName: i18n.T("ps.toggle.off"), Value: false},
			},
			SelectedOption: discoverExistingInstallsToIndex(config.ShouldDiscoverExistingInstalls()),
		},
		{
			Item:    gaba.MenuItem{Text: i18n.T("ps.settings.info")},
			Options: []gaba.Option{{Type: gaba.OptionTypeClickable}},
		},
	}
}

func (s *SettingsScreen) applySettings(config *internal.Config, items []gaba.ItemWithOptions) {
	for _, item := range items {
		switch item.Item.Text {
		case i18n.T("ps.settings.platform_filter"):
			if val, ok := item.Options[item.SelectedOption].Value.(internal.PlatformFilterMode); ok {
				config.PlatformFilter = val
			}
		case i18n.T("ps.settings.debug_level"):
			if val, ok := item.Options[item.SelectedOption].Value.(internal.DebugLevel); ok {
				config.DebugLevel = val
			}
		case i18n.T("ps.settings.discover"):
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
