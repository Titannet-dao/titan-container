package main

import (
	"fmt"
	"os"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/Filecoin-Titan/titan-container/build"
	lcli "github.com/Filecoin-Titan/titan-container/cli"
	cliutil "github.com/Filecoin-Titan/titan-container/cli/util"
	liblog "github.com/Filecoin-Titan/titan-container/lib/log"
	"github.com/Filecoin-Titan/titan-container/node"
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	"github.com/filecoin-project/go-jsonrpc"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var log = logging.Logger("main")

var AdvanceBlockCmd *cli.Command

const (
	// FlagManagerRepo Flag
	FlagManagerRepo = "manager-repo"
)

func main() {
	types.RunningNodeType = types.NodeManager

	liblog.SetupLogLevels()

	local := []*cli.Command{
		initCmd,
		runCmd,
	}

	if AdvanceBlockCmd != nil {
		local = append(local, AdvanceBlockCmd)
	}

	interactiveDef := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	app := &cli.App{
		Name:                 "manager",
		Usage:                "Titan Edge Cloud Computing Manager Service",
		Version:              build.UserVersion(),
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				// examined in the Before above
				Name:        "color",
				Usage:       "use color in display output",
				DefaultText: "depends on output being a TTY",
			},
			&cli.StringFlag{
				Name:    FlagManagerRepo,
				EnvVars: []string{"TITAN_MANAGER_PATH"},
				Hidden:  true,
				Value:   "~/.manager",
			},
			&cli.BoolFlag{
				Name:  "interactive",
				Usage: "setting to false will disable interactive functionality of commands",
				Value: interactiveDef,
			},
			&cli.BoolFlag{
				Name:  "force-send",
				Usage: "if true, will ignore pre-send checks",
			},
			cliutil.FlagVeryVerbose,
		},
		After: func(c *cli.Context) error {
			if r := recover(); r != nil {
				panic(r)
			}
			return nil
		},

		Commands: append(local, append(lcli.Commands, lcli.ManagerCMDs...)...),
	}

	app.Setup()
	app.Metadata["repoType"] = repo.Manager

	lcli.RunApp(app)
}

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialize a manager repo",
	Action: func(cctx *cli.Context) error {
		log.Info("Initializing manager service")
		repoPath := cctx.String(FlagManagerRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if ok {
			return xerrors.Errorf("repo at '%s' is already initialized", cctx.String(FlagManagerRepo))
		}

		if err := r.Init(repo.Manager); err != nil {
			return err
		}

		return nil
	},
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start manager service",
	Flags: []cli.Flag{},

	Before: func(cctx *cli.Context) error {
		return nil
	},
	Action: func(cctx *cli.Context) error {
		log.Info("Starting manager service")

		repoPath := cctx.String(FlagManagerRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if !ok {
			if err := r.Init(repo.Manager); err != nil {
				return err
			}
		}

		lr, err := r.Lock(repo.Manager)
		if err != nil {
			return err
		}

		cfg, err := lr.Config()
		if err != nil {
			return err
		}

		managerCfg := cfg.(*config.ManagerCfg)

		err = lr.Close()
		if err != nil {
			return err
		}

		shutdownChan := make(chan struct{})

		var managerAPI api.Manager
		stop, err := node.New(cctx.Context,
			node.Manager(&managerAPI),
			node.Base(),
			node.Repo(r),
		)
		if err != nil {
			return xerrors.Errorf("creating node: %w", err)
		}

		// Populate JSON-RPC options.
		serverOptions := []jsonrpc.ServerOption{jsonrpc.WithServerErrors(api.RPCErrors)}
		if maxRequestSize := cctx.Int("api-max-req-size"); maxRequestSize != 0 {
			serverOptions = append(serverOptions, jsonrpc.WithMaxRequestSize(int64(maxRequestSize)))
		}

		// Instantiate the manager handler.
		h, err := node.ManagerHandler(managerAPI, true, serverOptions...)
		if err != nil {
			return fmt.Errorf("failed to instantiate rpc handler: %s", err.Error())
		}

		// Serve the RPC.
		rpcStopper, err := node.ServeRPC(h, "manager", managerCfg.API.ListenAddress)
		if err != nil {
			return fmt.Errorf("failed to start json-rpc endpoint: %s", err.Error())
		}

		log.Info("manager listen with:", managerCfg.API.ListenAddress)

		// Monitor for shutdown.
		finishCh := node.MonitorShutdown(shutdownChan,
			node.ShutdownHandler{Component: "rpc server", StopFunc: rpcStopper},
			node.ShutdownHandler{Component: "node", StopFunc: stop},
		)
		<-finishCh // fires when shutdown is complete.
		return nil
	},
}
