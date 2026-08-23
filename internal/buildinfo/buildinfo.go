// Package buildinfo exposes release metadata injected by the build pipeline.
package buildinfo

import "runtime"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Info is the stable public representation of the running build.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Current returns the running build's metadata.
func Current() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		BuildDate: date,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
