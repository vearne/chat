package consts

import "time"

var (
	// Version logs build version injected with -ldflags -X opitons.
	Version string

	// BuildTime logs build time injected with -ldflags -X opitons.
	BuildTime string

	// GitTag logs git version and injected with -ldflags -X opitons.
	GitTag string

	// uptime
	UpTime string
)

func init() {
	UpTime = time.Now().Format(time.RFC3339)
}
