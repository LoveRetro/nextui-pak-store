package version

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

type BuildInfo struct {
	Version   string
	GitCommit string
	BuildDate string
}

func Get() BuildInfo {
	return BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
	}
}
