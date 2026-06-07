package processor

import "git.um-react.app/um/cli/algo/common"

type Config struct {
	InputDir        string
	OutputDir       string
	SkipNoop        bool
	RemoveSource    bool
	UpdateMetadata  bool
	OverwriteOutput bool

	Crypto common.CryptoParams
}
