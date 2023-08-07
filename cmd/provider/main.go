package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/Filecoin-Titan/titan-container/build"
	lcli "github.com/Filecoin-Titan/titan-container/cli"
	cliutil "github.com/Filecoin-Titan/titan-container/cli/util"
	liblog "github.com/Filecoin-Titan/titan-container/lib/log"
	"github.com/Filecoin-Titan/titan-container/metrics"
	"github.com/Filecoin-Titan/titan-container/node"
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	"github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"golang.org/x/xerrors"
)

var log = logging.Logger("main")
var HeartbeatInterval = 10 * time.Second

const (
	// FlagProviderRepo Flag
	FlagProviderRepo = "provider-repo"
)

var AdvanceBlockCmd *cli.Command

func main() {
	types.RunningNodeType = types.NodeProvider

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
		Name:                 "provider",
		Usage:                "Titan Edge Cloud Computing provider Service",
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
				Name:    FlagProviderRepo,
				EnvVars: []string{"TITAN_PROVIDER_PATH"},
				Hidden:  true,
				Value:   "~/.provider",
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

		Commands: append(local, lcli.Commands...),
	}

	app.Setup()
	app.Metadata["repoType"] = repo.Provider

	lcli.RunApp(app)
}

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialize a provider repo",
	Action: func(cctx *cli.Context) error {
		log.Info("Initializing provider service")
		repoPath := cctx.String(FlagProviderRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if ok {
			return xerrors.Errorf("repo at '%s' is already initialized", cctx.String(FlagProviderRepo))
		}

		if err := r.Init(repo.Provider); err != nil {
			return err
		}

		return nil
	},
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start provider service",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "http-server-timeout",
			Value: "30s",
		},
	},

	Before: func(cctx *cli.Context) error {
		return nil
	},
	Action: func(cctx *cli.Context) error {
		log.Info("Starting provider node")

		// Connect to manager
		ctx := lcli.ReqContext(cctx)

		var managerAPI api.Manager
		var err error
		var closer func()
		for {
			managerAPI, closer, err = lcli.GetManagerAPI(cctx)
			if err == nil {
				_, err = managerAPI.Version(ctx)
				if err == nil {
					break
				}
			}
			fmt.Printf("\r\x1b[0KConnecting to manager API... (%s)", err)
			time.Sleep(time.Second)
			continue
		}
		defer closer()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Register all metric views
		if err := view.Register(
			metrics.DefaultViews...,
		); err != nil {
			log.Fatalf("Cannot register the view: %v", err)
		}

		v, err := managerAPI.Version(ctx)
		if err != nil {
			return err
		}
		if v.APIVersion != api.ManagerAPIVersion0 {
			return xerrors.Errorf("manager API version doesn't match: expected: %s", api.APIVersion{APIVersion: api.ManagerAPIVersion0})
		}
		log.Infof("Remote version %s", v)

		repoPath := cctx.String(FlagProviderRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if !ok {
			if err := r.Init(repo.Provider); err != nil {
				return err
			}
		}

		lr, err := r.Lock(repo.Provider)
		if err != nil {
			return err
		}

		cfg, err := lr.Config()
		if err != nil {
			return err
		}

		providerID, err := r.UUID()
		if err != nil {
			if err == repo.ErrNoUUID {
				providerID, err = uuid.New().MarshalText()
				if err != nil {
					return err
				}
				err = lr.SetUUID(providerID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		providerCfg := cfg.(*config.ProviderCfg)

		err = lr.Close()
		if err != nil {
			return err
		}

		var providerAPI api.Provider
		stop, err := node.New(cctx.Context,
			node.Provider(&providerAPI),
			node.Base(),
			node.Repo(r),
			node.Override(new(api.Manager), managerAPI),
		)
		if err != nil {
			return xerrors.Errorf("creating node: %w", err)
		}

		address := providerCfg.API.ListenAddress
		log.Info("Setting up control endpoint at " + address)

		timeout, err := time.ParseDuration(cctx.String("http-server-timeout"))
		if err != nil {
			return xerrors.Errorf("invalid time string %s: %x", cctx.String("http-server-timeout"), err)
		}

		srv := &http.Server{
			Handler:           node.ProviderHandler(managerAPI.AuthVerify, providerAPI, true),
			ReadHeaderTimeout: timeout,
			BaseContext: func(listener net.Listener) context.Context {
				ctx, _ := tag.New(context.Background(), tag.Upsert(metrics.APIInterface, "provider"))
				return ctx
			},
		}

		go func() {
			<-ctx.Done()
			log.Warn("Shutting down...")
			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Errorf("shutting down RPC server failed: %s", err)
			}
			stop(ctx) //nolint:errcheck
			log.Warn("Graceful shutdown successful")
		}()

		nl, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		managerSession, err := managerAPI.Session(ctx)
		if err != nil {
			return xerrors.Errorf("getting manager session: %w", err)
		}

		waitQuietCh := func() chan struct{} {
			out := make(chan struct{})
			go func() {
				close(out)
			}()
			return out
		}

		go func() {
			heartbeats := time.NewTicker(HeartbeatInterval)
			defer heartbeats.Stop()

			var readyCh chan struct{}
			for {

				// TODO: we could get rid of this, but that requires tracking resources for restarted tasks correctly
				if readyCh == nil {
					log.Info("Making sure no local tasks are running")
					readyCh = waitQuietCh()
				}

				for {
					curSession, err := managerAPI.Session(ctx)
					if err != nil {
						log.Errorf("heartbeat: checking remote session failed: %+v", err)
					} else {
						if curSession != managerSession {
							managerSession = curSession
							break
						}
					}

					select {
					case <-readyCh:
						if err := managerAPI.ProviderConnect(ctx, "http://"+address+"/rpc/v0",
							&types.Provider{
								ID:      types.ProviderID(providerID),
								Owner:   providerCfg.Owner,
								HostURI: providerCfg.HostURI,
							}); err != nil {
							log.Errorf("Registering provider failed: %+v", err)
							cancel()
							return
						}

						log.Info("Provider registered successfully, waiting for tasks")

						readyCh = nil
					case <-heartbeats.C:
					case <-ctx.Done():
						return // graceful shutdown
					}
				}

				log.Errorf("MANAGER CONNECTION LOST")
			}
		}()

		return srv.Serve(nl)
	},
}
