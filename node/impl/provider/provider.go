package provider

import (
	"context"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

var session = uuid.New()

// Provider represents a provider service in a cloud computing system.
type Provider struct {
	fx.In

	Manager Manager
}

var _ api.Provider = &Provider{}

func (p *Provider) Session(ctx context.Context) (uuid.UUID, error) {
	return session, nil
}

func (p *Provider) Version(context.Context) (api.Version, error) {
	return api.ProviderAPIVersion0, nil
}

func (p *Provider) GetStatistics(ctx context.Context) (*types.ResourcesStatistics, error) {
	return p.Manager.GetStatistics(ctx)
}

func (p *Provider) GetDeployment(ctx context.Context, id types.DeploymentID) (*types.Deployment, error) {
	return p.Manager.GetDeployment(ctx, id)
}

func (p *Provider) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.Manager.CreateDeployment(ctx, deployment)
}

func (p *Provider) UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.Manager.UpdateDeployment(ctx, deployment)
}

func (p *Provider) CloseDeployment(ctx context.Context, deployment *types.Deployment) error {
	return p.Manager.CloseDeployment(ctx, deployment)
}

func (p *Provider) GetLogs(ctx context.Context, id types.DeploymentID) ([]*types.ServiceLog, error) {
	return p.Manager.GetLogs(ctx, id)
}
func (p *Provider) GetEvents(ctx context.Context, id types.DeploymentID) ([]*types.ServiceEvent, error) {
	return p.Manager.GetEvents(ctx, id)
}
