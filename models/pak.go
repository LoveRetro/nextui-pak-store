package models

import (
	"qlova.tech/sum"
)

type Pak struct {
	ID               string            `json:"id"`
	StorefrontName   string            `json:"storefront_name"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	PakType          sum.Int[PakType]  `json:"type"`
	Description      string            `json:"description"`
	Author           string            `json:"author"`
	RepoURL          string            `json:"repo_url"`
	PreviousRepoURLs []string          `json:"previous_repo_urls"`
	ReleaseFilename  string            `json:"release_filename"`
	Changelog        map[string]string `json:"changelog"`
	PreviousNames    []string          `json:"previous_names"`
	Scripts          Scripts           `json:"scripts"`
	UpdateIgnore     []string          `json:"update_ignore"`
	Screenshots      []string          `json:"screenshots"`
	Platforms        []string          `json:"platforms"`
	Categories       []string          `json:"categories"`
	LargePak         bool              `json:"large_pak"`
	Disabled         bool              `json:"disabled"`

	IsPakZ       bool `json:"-"`
	CanUninstall bool `json:"-"`
}

type Scripts struct {
	PostInstall   Script `json:"post_install"`
	PostUpdate    Script `json:"post_update"`
	PostUninstall Script `json:"post_uninstall"`
}

type Script struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

type PakType struct {
	TOOL,
	EMU sum.Int[PakType]
}

var PakTypeMap = map[sum.Int[PakType]]string{
	PakTypes.TOOL: "TOOL",
	PakTypes.EMU:  "EMU",
}

var PakTypes = sum.Int[PakType]{}.Sum()

func (p Pak) Value() interface{} {
	return p
}

func (p Pak) HasScripts() bool {
	return p.Scripts.PostInstall.Path != "" || p.Scripts.PostUpdate.Path != ""
}
