package cli

import (
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"

	"github.com/Filecoin-Titan/titan-container/api"
	cliutil "github.com/Filecoin-Titan/titan-container/cli/util"
)

var log = logging.Logger("cli")

// custom CLI error

type ErrCmdFailed struct {
	msg string
}

func (e *ErrCmdFailed) Error() string {
	return e.msg
}

func NewCliError(s string) error {
	return &ErrCmdFailed{s}
}

// APIConnector returns API instance
type APIConnector func() api.Manager

var (
	GetAPIInfo = cliutil.GetAPIInfo
	GetRawAPI  = cliutil.GetRawAPI
	GetAPI     = cliutil.GetCommonAPI
)

var (
	DaemonContext = cliutil.DaemonContext
	ReqContext    = cliutil.ReqContext
)

var (
	GetManagerAPI  = cliutil.GetManagerAPI
	GetProviderAPI = cliutil.GetProviderAPI
)

var CommonCommands = []*cli.Command{
	AuthCmd,
	LogCmd,
	PprofCmd,
	VersionCmd,
}

var Commands = []*cli.Command{
	WithCategory("developer", AuthCmd),
	WithCategory("developer", LogCmd),
	PprofCmd,
	VersionCmd,
}

func WithCategory(cat string, cmd *cli.Command) *cli.Command {
	cmd.Category = strings.ToUpper(cat)
	return cmd
}
