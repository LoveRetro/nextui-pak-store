package ui

import (
	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	icons "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
)

func FooterSelect() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "Select"}
}

func FooterView() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "View"}
}

func FooterConfirm() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "Confirm"}
}

func FooterBack() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: "Back"}
}

func FooterQuit() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: "Quit"}
}

func FooterCancel() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: "Cancel"}
}

func FooterInstall() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "Install"}
}

func FooterUpdate() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "Update"}
}

func FooterUninstall() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: "Uninstall"}
}

func BackSelectFooter() []gaba.FooterHelpItem {
	return []gaba.FooterHelpItem{FooterBack(), FooterSelect()}
}

func BackViewFooter() []gaba.FooterHelpItem {
	return []gaba.FooterHelpItem{FooterBack(), FooterView()}
}

func QuitSelectFooter() []gaba.FooterHelpItem {
	return []gaba.FooterHelpItem{FooterQuit(), FooterSelect()}
}

func OptionsListFooter() []gaba.FooterHelpItem {
	return []gaba.FooterHelpItem{
		FooterCancel(),
		{ButtonName: icons.LeftRight, HelpText: "Cycle"},
		{ButtonName: "Start", HelpText: "Save"},
	}
}
