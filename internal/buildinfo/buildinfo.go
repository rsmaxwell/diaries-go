package buildinfo

type BuildInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GitCommit string `json:"gitCommit"`
	GitBranch string `json:"gitBranch"`
	GitURL    string `json:"gitUrl"`
}

func NewBuildInfo() *BuildInfo {

	info := new(BuildInfo)
	info.Version = "<BUILD_ID>"
	info.BuildDate = "<BUILD_DATE>"
	info.GitCommit = "<GIT_COMMIT>"
	info.GitBranch = "<GIT_BRANCH>"
	info.GitURL = "<GIT_URL>"

	return info
}
