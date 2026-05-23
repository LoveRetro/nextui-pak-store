package ui

import (
	"errors"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/LoveRetro/nextui-pak-store/internal/i18n"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/utils"
	"github.com/LoveRetro/nextui-pak-store/version"
)

type InfoInput struct{}

type InfoOutput struct{}

type InfoScreen struct{}

func NewInfoScreen() *InfoScreen {
	return &InfoScreen{}
}

func (s *InfoScreen) Draw(input InfoInput) (ScreenResult[InfoOutput], error) {
	output := InfoOutput{}

	sections := s.buildSections()

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ShowScrollbar = true

	_, err := gaba.DetailScreen("", options, []gaba.FooterHelpItem{
		FooterBack(),
	})

	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		gaba.GetLogger().Error("Info screen error", "error", err)
		return withAction(output, ActionError), err
	}

	return back(output), nil
}

func (s *InfoScreen) buildSections() []gaba.Section {
	sections := make([]gaba.Section, 0)

	buildInfo := version.Get()
	buildMetadata := []gaba.MetadataItem{
		{Label: i18n.T("ps.info.version"), Value: buildInfo.Version},
		{Label: i18n.T("ps.info.commit"), Value: buildInfo.GitCommit},
		{Label: i18n.T("ps.info.build_date"), Value: buildInfo.BuildDate},
	}
	sections = append(sections, gaba.NewInfoSection("Pak Store", buildMetadata))

	sections = append(sections, gaba.NewDescriptionSection(
		i18n.T("ps.info.community.title"),
		i18n.T("ps.info.community.body"),
	))

	qrcode, err := utils.CreateTempQRCode(models.PakStoreRepo, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			i18n.T("ps.info.github_repo"),
			qrcode,
			int32(256),
			int32(256),
			constants.TextAlignCenter,
		))
	} else {
		gaba.GetLogger().Error("Unable to generate QR code for repository", "error", err)
	}

	return sections
}
