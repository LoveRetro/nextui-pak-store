package ui

import (
	"errors"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
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
		{Label: "Version", Value: buildInfo.Version},
		{Label: "Commit", Value: buildInfo.GitCommit},
		{Label: "Build Date", Value: buildInfo.BuildDate},
	}
	sections = append(sections, gaba.NewInfoSection("Pak Store", buildMetadata))

	sections = append(sections, gaba.NewDescriptionSection(
		"Community Shout Out",
		"Pak Store exists because of the incredible NextUI community. "+
			"Your creativity, passion, and dedication to building amazing paks "+
			"is what makes this platform special. Every emulator, tool, and enhancement "+
			"you create brings joy to our retro doo-dads! "+
			"Thank you for sharing your talents and making NextUI better for everyone.",
	))

	qrcode, err := utils.CreateTempQRCode(models.PakStoreRepo, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"GitHub Repository",
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
