package common

import "fmt"

const (
	MajorVersion = 2
	MinorVersion = 5
	FixVersion   = 0
	CommitHash   = ""
	VERSION      = "2.0.0"
)

func GetVersion() string {
	if CommitHash != "" {
		return fmt.Sprintf("swan-miner-v%v.%v.%v-%s", MajorVersion, MinorVersion, FixVersion, CommitHash)
	} else {
		return fmt.Sprintf("swan-miner-v%v.%v.%v", MajorVersion, MinorVersion, FixVersion)
	}
}
