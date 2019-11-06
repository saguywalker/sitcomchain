package version

var (
	// GitCommit is the current HEAD set using ldflags.
	GitCommit string

	// Version is the built ABCI app version.
	Version string = ABCIAppSemVer

	// AppProtocolVersion is ABCI App protocol version.
	AppProtocolVersion uint64 = ABCIAppProtocolVersion
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}

const (
	// ABCIAppSemVer is ABCI app version.
	ABCIAppSemVer = "1.0.0"

	// ABCIAppProtocolVersion is ABCI App protocol version.
	ABCIAppProtocolVersion = 1
)
