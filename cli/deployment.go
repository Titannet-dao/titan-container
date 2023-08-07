package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/Filecoin-Titan/titan-container/lib/tablewriter"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/yaml"
)

var defaultDateTimeLayout = "2006-01-02 15:04:05"

var deploymentCmds = &cli.Command{
	Name:  "deployment",
	Usage: "Manager deployment",
	Subcommands: []*cli.Command{
		CreateDeployment,
		DeploymentList,
		DeleteDeployment,
		StatusDeployment,
	},
}

var CreateDeployment = &cli.Command{
	Name:  "create",
	Usage: "create new deployment",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "provider-id",
			Usage:    "the provider id",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "owner",
			Usage: "owner address",
		},
		&cli.StringFlag{
			Name:  "template",
			Usage: "from the template file",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "deployment name",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "deployment internal server port",
		},
		&cli.BoolFlag{
			Name:  "auth",
			Usage: "deploy from authority",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "deployment image",
		},
		&cli.Float64Flag{
			Name:  "cpu",
			Usage: "cpu cores",
		},
		&cli.Int64Flag{
			Name:  "mem",
			Usage: "memory",
		},
		&cli.Int64Flag{
			Name:  "storage",
			Usage: "storage",
		},
		&cli.StringFlag{
			Name:  "env",
			Usage: "set the deployment running environment",
		},
		&cli.StringSliceFlag{
			Name:  "args",
			Usage: "set the deployment running arguments",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)
		providerID := types.ProviderID(cctx.String("provider-id"))

		if cctx.String("template") != "" {
			return createDeploymentFromTemplate(ctx, api, providerID, cctx.String("template"))
		}

		if cctx.String("image") == "" {
			return errors.Errorf("Required flags image not set")
		}

		var env types.Env
		if cctx.String("env") != "" {
			err := json.Unmarshal([]byte(cctx.String("env")), &env)
			if err != nil {
				return err
			}
		}

		deployment := &types.Deployment{
			ProviderID: providerID,
			Name:       cctx.String("name"),
			Authority:  cctx.Bool("auth"),
			Services: []*types.Service{
				{
					Image: cctx.String("image"),
					Ports: []types.Port{
						{
							Port: cctx.Int("port"),
						},
					},
					ComputeResources: types.ComputeResources{
						CPU:     cctx.Float64("cpu"),
						Memory:  cctx.Int64("mem"),
						Storage: cctx.Int64("storage"),
					},
					Env:       env,
					Arguments: cctx.StringSlice("args"),
				},
			},
		}

		return api.CreateDeployment(ctx, deployment)
	},
}

func createDeploymentFromTemplate(ctx context.Context, api api.Manager, providerID types.ProviderID, path string) error {
	yamlFiles, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var deployment types.Deployment
	err = yaml.Unmarshal(yamlFiles, &deployment)
	if err != nil {
		return err
	}

	deployment.ProviderID = providerID
	return api.CreateDeployment(ctx, &deployment)
}

var DeploymentList = &cli.Command{
	Name:  "list",
	Usage: "List deployments",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "owner",
			Usage: "owner address",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "the deployment id",
		},
		&cli.BoolFlag{
			Name:  "show-all",
			Usage: "show deleted and inactive deployments",
		},
		&cli.IntFlag{
			Name:  "page",
			Usage: "the page number",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "size",
			Usage: "the page size",
			Value: 10,
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		tw := tablewriter.New(
			tablewriter.Col("ID"),
			tablewriter.Col("Image"),
			tablewriter.Col("State"),
			tablewriter.Col("Authority"),
			tablewriter.Col("Total"),
			tablewriter.Col("Ready"),
			tablewriter.Col("Available"),
			tablewriter.Col("CPU"),
			tablewriter.Col("Memory"),
			tablewriter.Col("Storage"),
			tablewriter.Col("Provider"),
			tablewriter.Col("Port"),
			tablewriter.Col("CreatedTime"),
		)

		opts := &types.GetDeploymentOption{
			Owner:        cctx.String("owner"),
			State:        []types.DeploymentState{types.DeploymentStateActive},
			DeploymentID: types.DeploymentID(cctx.String("id")),
			Page:         cctx.Int("page"),
			Size:         cctx.Int("size"),
		}

		if cctx.Bool("show-all") {
			opts.State = types.AllDeploymentStates
		}

		deployments, err := api.GetDeploymentList(ctx, opts)
		if err != nil {
			return err
		}

		for _, deployment := range deployments {
			for _, service := range deployment.Services {
				state := types.DeploymentStateInActive
				if service.Status.TotalReplicas == service.Status.ReadyReplicas {
					state = types.DeploymentStateActive
				}

				var exposePorts []string
				for _, port := range service.Ports {
					exposePorts = append(exposePorts, fmt.Sprintf("%d->%d", port.Port, port.ExposePort))
				}

				m := map[string]interface{}{
					"ID":          deployment.ID,
					"Image":       service.Image,
					"State":       types.DeploymentStateString(state),
					"Authority":   deployment.Authority,
					"Total":       service.Status.TotalReplicas,
					"Ready":       service.Status.ReadyReplicas,
					"Available":   service.Status.AvailableReplicas,
					"CPU":         service.CPU,
					"Memory":      units.BytesSize(float64(service.Memory * units.MiB)),
					"Storage":     units.BytesSize(float64(service.Storage * units.MiB)),
					"Provider":    deployment.ProviderExposeIP,
					"Port":        strings.Join(exposePorts, " "),
					"CreatedTime": deployment.CreatedAt.Format(defaultDateTimeLayout),
				}
				tw.Write(m)
			}
		}

		tw.Flush(os.Stdout)
		return nil
	},
}

var DeleteDeployment = &cli.Command{
	Name:  "delete",
	Usage: "delete deployment",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 1 {
			return IncorrectNumArgs(cctx)
		}

		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)
		deploymentID := types.DeploymentID(cctx.Args().First())

		deployments, err := api.GetDeploymentList(ctx, &types.GetDeploymentOption{
			DeploymentID: deploymentID,
		})
		if err != nil {
			return err
		}

		if len(deployments) == 0 {
			return errors.New("deployment not found")
		}

		for _, deployment := range deployments {
			err = api.CloseDeployment(ctx, deployment)
			if err != nil {
				log.Errorf("delete deployment failed: %v", err)
			}
		}

		return nil
	},
}

var StatusDeployment = &cli.Command{
	Name:  "status",
	Usage: "show deployment status",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "log",
			Usage: "show deployment log",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 1 {
			return IncorrectNumArgs(cctx)
		}

		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)
		deploymentID := types.DeploymentID(cctx.Args().First())

		deployments, err := api.GetDeploymentList(ctx, &types.GetDeploymentOption{
			DeploymentID: deploymentID,
		})
		if err != nil {
			return err
		}

		var deployment *types.Deployment
		for _, d := range deployments {
			if d.ID == deploymentID {
				deployment = d
				continue
			}
		}

		if deployment == nil {
			return errors.New("deployment not found")
		}

		fmt.Printf("DeploymentID:\t%s\n", deployment.ID)
		fmt.Printf("State:\t\t%s\n", types.DeploymentStateString(deployment.State))
		fmt.Printf("CreadTime:\t%v\n", deployment.CreatedAt)
		fmt.Printf("--------\nEvents:\n")

		serviceEvents, err := api.GetEvents(ctx, deployment)
		if err != nil {
			return err
		}

		for _, sv := range serviceEvents {
			for i, event := range sv.Events {
				fmt.Printf("%d.\t[%s]\t%s\n", i, sv.ServiceName, event)
			}
		}

		if cctx.Bool("log") {
			fmt.Printf("--------\nLogs:\n")

			serviceLogs, err := api.GetLogs(ctx, deployment)
			if err != nil {
				return err
			}

			for _, sl := range serviceLogs {
				for _, l := range sl.Logs {
					fmt.Printf("%s\n", l)
				}
			}
		}

		return nil
	},
}
