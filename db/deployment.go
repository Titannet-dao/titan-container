package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/jmoiron/sqlx"
)

func (m *ManagerDB) CreateDeployment(ctx context.Context, deployment *types.Deployment) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addNewDeployment(ctx, tx, deployment)
	if err != nil {
		return err
	}

	err = addNewServices(ctx, tx, deployment.Services)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func addNewDeployment(ctx context.Context, tx *sqlx.Tx, deployment *types.Deployment) error {
	qry := `INSERT INTO deployments (id, name, owner, state, type, authority, version, balance, cost, expiration, provider_id, created_at, updated_at) 
		        VALUES (:id, :name, :owner, :state, :type, :authority, :version, :balance, :cost, :expiration, :provider_id, :created_at, :updated_at)
		         ON DUPLICATE KEY UPDATE  state=:state, authority=:authority, version=:version, balance=:balance, cost=:cost, expiration=:expiration, updated_at=:updated_at`
	_, err := tx.NamedExecContext(ctx, qry, deployment)

	return err
}

func addNewServices(ctx context.Context, tx *sqlx.Tx, services []*types.Service) error {
	qry := `INSERT INTO services (id, name, image, ports, cpu, memory, storage, deployment_id, env, arguments, error_message, created_at, updated_at) 
		        VALUES (:id,:name, :image, :ports, :cpu, :memory, :storage, :deployment_id, :env, :arguments, :error_message, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, qry, services)

	return err
}

type DeploymentService struct {
	types.Deployment
	types.Service `db:"service"`
}

func (m *ManagerDB) GetDeployments(ctx context.Context, option *types.GetDeploymentOption) ([]*types.Deployment, error) {
	var ds []*DeploymentService
	qry := `SELECT d.*, s.image as 'service.image', 
			s.name as 'service.name',
			s.cpu as 'service.cpu', 
			s.memory as 'service.memory',
			s.storage as 'service.storage', 
			s.ports as 'service.ports', 
			s.env as 'service.env', 
			s.arguments as 'service.arguments', 
			s.error_message  as 'service.error_message',
			p.host_uri  as 'provider_expose_ip'
		FROM deployments d LEFT JOIN services s ON d.id = s.deployment_id LEFT JOIN providers p ON d.provider_id = p.id`

	var condition []string
	if option.DeploymentID != "" {
		condition = append(condition, fmt.Sprintf(`d.id = '%s'`, option.DeploymentID))
	}

	if option.Owner != "" {
		condition = append(condition, fmt.Sprintf(`d.owner = '%s'`, option.Owner))
	}

	if len(option.State) > 0 {
		var states []string
		for _, s := range option.State {
			states = append(states, strconv.Itoa(int(s)))
		}
		condition = append(condition, fmt.Sprintf(`d.state in (%s)`, strings.Join(states, ",")))
	}

	if len(condition) > 0 {
		qry += ` WHERE `
		qry += strings.Join(condition, ` AND `)
	}

	if option.Page <= 0 {
		option.Page = 1
	}

	if option.Size <= 0 {
		option.Size = 10
	}

	offset := (option.Page - 1) * option.Size
	limit := option.Size
	qry += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	err := m.db.SelectContext(ctx, &ds, qry)
	if err != nil {
		return nil, err
	}

	var out []*types.Deployment
	deploymentToServices := make(map[types.DeploymentID]*types.Deployment)
	for _, d := range ds {
		_, ok := deploymentToServices[d.Deployment.ID]
		if !ok {
			deploymentToServices[d.Deployment.ID] = &d.Deployment
			deploymentToServices[d.Deployment.ID].Services = make([]*types.Service, 0)
			out = append(out, deploymentToServices[d.Deployment.ID])
		}
		deploymentToServices[d.Deployment.ID].Services = append(deploymentToServices[d.Deployment.ID].Services, &d.Service)
	}

	return out, nil
}

func (m *ManagerDB) UpdateDeploymentState(ctx context.Context, id types.DeploymentID, state types.DeploymentState) error {
	qry := `Update deployments set state = ? where id = ?`
	_, err := m.db.ExecContext(ctx, qry, state, id)
	return err
}

func (m *ManagerDB) AddProperties(ctx context.Context, properties *types.Properties) error {
	qry := `INSERT INTO properties (id, provider_id, app_id, app_type, created_at, updated_at) 
		        VALUES (:id, :provider_id, :app_id, :app_type, :created_at, :updated_at) ON DUPLICATE KEY UPDATE 
		        app_id=:app_id, app_type=:app_type, updated_at=:updated_at`
	_, err := m.db.NamedExecContext(ctx, qry, properties)

	return err
}
