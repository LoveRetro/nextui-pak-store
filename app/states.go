package main

import (
	"os"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/router"
	"github.com/LoveRetro/nextui-pak-store/internal"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/ui"
)

const (
	screenMainMenu router.Screen = iota
	screenBrowse
	screenPakList
	screenPakInfo
	screenUpdates
	screenManageInstalled
	screenSettings
	screenInfo
)

type ListPosition struct {
	Index             int
	VisibleStartIndex int
}

type BrowseResume struct {
	Pos ListPosition
}

type PakListResume struct {
	Pos      ListPosition
	Category string
}

type UpdatesResume struct {
	Pos ListPosition
}

type ManageResume struct {
	Pos ListPosition
}

type BrowseInputWithResume struct {
	Storefront models.Storefront
	Resume     *BrowseResume
}

type PakListInputWithResume struct {
	Storefront models.Storefront
	Category   string
	Resume     *PakListResume
}

type UpdatesInputWithResume struct {
	Storefront models.Storefront
	Resume     *UpdatesResume
}

type ManageInputWithResume struct {
	Storefront models.Storefront
	Resume     *ManageResume
}

type PakInfoInputWithSource struct {
	Paks        []models.Pak
	Category    string
	IsUpdate    bool
	IsInstalled bool
	Source      router.Screen
}

func buildRouter(storefront models.Storefront) *router.Router {
	r := router.New()

	r.Register(screenMainMenu, func(input any) (any, error) {
		screen := ui.NewMainMenuScreen()
		result, err := screen.Draw(ui.MainMenuInput{
			Storefront: storefront,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenBrowse, func(input any) (any, error) {
		in := input.(BrowseInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewBrowseScreen()
		result, err := screen.Draw(ui.BrowseInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenPakList, func(input any) (any, error) {
		in := input.(PakListInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewPakListScreen()
		result, err := screen.Draw(ui.PakListInput{
			Storefront:           in.Storefront,
			Category:             in.Category,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenPakInfo, func(input any) (any, error) {
		in := input.(PakInfoInputWithSource)

		screen := ui.NewPakInfoScreen()
		result, err := screen.Draw(ui.PakInfoInput{
			Paks:        in.Paks,
			Category:    in.Category,
			IsUpdate:    in.IsUpdate,
			IsInstalled: in.IsInstalled,
		})
		if err != nil {
			return result, err
		}

		return struct {
			Result      ui.ScreenResult[ui.PakInfoOutput]
			Source      router.Screen
			Paks        []models.Pak
			Category    string
			IsUpdate    bool
			IsInstalled bool
		}{result, in.Source, in.Paks, in.Category, in.IsUpdate, in.IsInstalled}, nil
	})

	r.Register(screenUpdates, func(input any) (any, error) {
		in := input.(UpdatesInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewUpdatesScreen()
		result, err := screen.Draw(ui.UpdatesInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenManageInstalled, func(input any) (any, error) {
		in := input.(ManageInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewManageInstalledScreen()
		result, err := screen.Draw(ui.ManageInstalledInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenSettings, func(input any) (any, error) {
		config := internal.GetConfig()
		screen := ui.NewSettingsScreen()
		result, err := screen.Draw(ui.SettingsInput{
			Config: config,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.Register(screenInfo, func(input any) (any, error) {
		screen := ui.NewInfoScreen()
		result, err := screen.Draw(ui.InfoInput{})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	r.OnTransition(func(from router.Screen, result any, stack *router.Stack) (router.Screen, any) {
		switch from {
		case screenMainMenu:
			r := result.(ui.ScreenResult[ui.MainMenuOutput])
			switch r.Action {
			case ui.ActionBrowse:
				return screenBrowse, BrowseInputWithResume{Storefront: storefront}
			case ui.ActionUpdates:
				return screenUpdates, UpdatesInputWithResume{Storefront: storefront}
			case ui.ActionManageInstalled:
				return screenManageInstalled, ManageInputWithResume{Storefront: storefront}
			case ui.ActionSettings:
				return screenSettings, nil
			case ui.ActionQuit, ui.ActionError:
				return router.ScreenExit, nil
			}

		case screenBrowse:
			r := result.(ui.ScreenResult[ui.BrowseOutput])
			switch r.Action {
			case ui.ActionSelected:
				stack.Push(from, BrowseInputWithResume{Storefront: storefront}, &BrowseResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakList, PakListInputWithResume{
					Storefront: storefront,
					Category:   r.Value.SelectedCategory,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}

		case screenPakList:
			r := result.(ui.ScreenResult[ui.PakListOutput])
			switch r.Action {
			case ui.ActionSelected:
				stack.Push(from, PakListInputWithResume{
					Storefront: storefront,
					Category:   r.Value.Category,
				}, &PakListResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
					Category: r.Value.Category,
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        []models.Pak{r.Value.SelectedPak},
					Category:    r.Value.Category,
					IsUpdate:    r.Value.HasUpdate,
					IsInstalled: r.Value.IsInstalled,
					Source:      screenPakList,
				}
			case ui.ActionBack:
				if entry := stack.Pop(); entry != nil {
					in := entry.Input.(BrowseInputWithResume)
					if entry.Resume != nil {
						in.Resume = entry.Resume.(*BrowseResume)
					}
					return screenBrowse, in
				}
				return screenBrowse, BrowseInputWithResume{Storefront: storefront}
			}

		case screenPakInfo:
			wrapper := result.(struct {
				Result      ui.ScreenResult[ui.PakInfoOutput]
				Source      router.Screen
				Paks        []models.Pak
				Category    string
				IsUpdate    bool
				IsInstalled bool
			})
			r := wrapper.Result
			source := wrapper.Source

			switch r.Action {
			case ui.ActionInstallSuccess:
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        wrapper.Paks,
					Category:    wrapper.Category,
					IsUpdate:    false,
					IsInstalled: true,
					Source:      source,
				}

			case ui.ActionPakStoreUpdated:
				gaba.ProcessMessage("Pak Store Updated! Exiting...", gaba.ProcessMessageOptions{}, func() (any, error) {
					time.Sleep(3 * time.Second)
					return nil, nil
				})
				os.Exit(0)
				return router.ScreenExit, nil

			case ui.ActionUninstalled:
				switch source {
				case screenManageInstalled:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(ManageInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*ManageResume)
						}
						return screenManageInstalled, in
					}
					return screenManageInstalled, ManageInputWithResume{Storefront: storefront}

				case screenUpdates:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(UpdatesInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*UpdatesResume)
						}
						return screenUpdates, in
					}
					return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

				default: // screenPakList - return to pak info showing as uninstalled
					return screenPakInfo, PakInfoInputWithSource{
						Paks:        wrapper.Paks,
						Category:    wrapper.Category,
						IsUpdate:    false,
						IsInstalled: false,
						Source:      source,
					}
				}

			case ui.ActionPartialUpdate:
				if entry := stack.Pop(); entry != nil {
					in := entry.Input.(UpdatesInputWithResume)
					if entry.Resume != nil {
						in.Resume = entry.Resume.(*UpdatesResume)
					}
					return screenUpdates, in
				}
				return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

			case ui.ActionCancelled, ui.ActionError:
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        wrapper.Paks,
					Category:    wrapper.Category,
					IsUpdate:    wrapper.IsUpdate,
					IsInstalled: wrapper.IsInstalled,
					Source:      source,
				}

			case ui.ActionBack, ui.ActionSelected:
				switch source {
				case screenManageInstalled:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(ManageInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*ManageResume)
						}
						return screenManageInstalled, in
					}
					return screenManageInstalled, ManageInputWithResume{Storefront: storefront}

				case screenUpdates:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(UpdatesInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*UpdatesResume)
						}
						return screenUpdates, in
					}
					return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

				default: // screenPakList
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(PakListInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*PakListResume)
						}
						return screenPakList, in
					}
					return screenPakList, PakListInputWithResume{Storefront: storefront}
				}
			}

		case screenUpdates:
			r := result.(ui.ScreenResult[ui.UpdatesOutput])
			switch r.Action {
			case ui.ActionSelected:
				stack.Push(from, UpdatesInputWithResume{Storefront: storefront}, &UpdatesResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:     r.Value.SelectedPaks,
					IsUpdate: true,
					Source:   screenUpdates,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}

		case screenManageInstalled:
			r := result.(ui.ScreenResult[ui.ManageInstalledOutput])
			switch r.Action {
			case ui.ActionSelected:
				stack.Push(from, ManageInputWithResume{Storefront: storefront}, &ManageResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        []models.Pak{r.Value.SelectedPak},
					IsUpdate:    false,
					IsInstalled: true,
					Source:      screenManageInstalled,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}

		case screenSettings:
			r := result.(ui.ScreenResult[ui.SettingsOutput])
			switch r.Action {
			case ui.ActionBack, ui.ActionSettingsSaved:
				return screenMainMenu, nil
			case ui.ActionInfo:
				return screenInfo, nil
			}

		case screenInfo:
			r := result.(ui.ScreenResult[ui.InfoOutput])
			switch r.Action {
			case ui.ActionBack:
				return screenSettings, nil
			}
		}

		return router.ScreenExit, nil
	})

	return r
}

func runApp(storefront models.Storefront) error {
	r := buildRouter(storefront)
	return r.Run(screenMainMenu, nil)
}
