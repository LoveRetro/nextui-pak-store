package models

// ScreenName constants for backwards compatibility
type ScreenName struct {
	MainMenu,
	Browse,
	PakList,
	PakInfo,
	DownloadPak,
	Updates,
	ManageInstalled int
}

// Legacy screen names - kept for any code that might still reference them
var ScreenNames = ScreenName{
	MainMenu:        0,
	Browse:          1,
	PakList:         2,
	PakInfo:         3,
	DownloadPak:     4,
	Updates:         5,
	ManageInstalled: 6,
}
