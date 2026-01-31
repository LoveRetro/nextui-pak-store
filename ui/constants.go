package ui

type Action int

const (
	ActionNone Action = iota
	ActionBack
	ActionQuit
	ActionSelected
	ActionError

	ActionBrowse
	ActionUpdates
	ActionManageInstalled
	ActionSettings
	ActionInfo

	ActionPakStoreUpdated
	ActionUninstalled
	ActionPartialUpdate
	ActionCancelled
	ActionInstallSuccess

	ActionSettingsSaved
	ActionDiscoverExistingInstalls
)
