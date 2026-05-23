package ui

import (
	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	icons "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/internal/i18n"
)

func FooterSelect() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.select")}
}

func FooterView() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.view")}
}

func FooterConfirm() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.confirm")}
}

func FooterBack() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: i18n.T("ps.btn.back")}
}

func FooterQuit() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: i18n.T("ps.btn.quit")}
}

func FooterCancel() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "B", HelpText: i18n.T("ps.btn.cancel")}
}

func FooterInstall() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.install")}
}

func FooterUpdate() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.update")}
}

func FooterUninstall() gaba.FooterHelpItem {
	return gaba.FooterHelpItem{ButtonName: "A", HelpText: i18n.T("ps.btn.uninstall")}
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
		{ButtonName: icons.LeftRight, HelpText: i18n.T("ps.btn.cycle")},
		{ButtonName: "Start", HelpText: i18n.T("ps.btn.save")},
	}
}
