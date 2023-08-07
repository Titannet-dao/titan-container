package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Filecoin-Titan/titan-container/api/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ManagerDB struct {
	db *sqlx.DB
}

func NewManagerDB(db *sqlx.DB) *ManagerDB {
	return &ManagerDB{
		db: db,
	}
}

func (m *ManagerDB) AddNewProvider(ctx context.Context, provider *types.Provider) error {
	qry := `INSERT INTO providers (id, owner, host_uri, ip, state, created_at, updated_at) 
		        VALUES (:id, :owner, :host_uri, :ip, :state, :created_at, :updated_at) ON DUPLICATE KEY UPDATE  owner=:owner, host_uri=:host_uri, 
		            ip=:ip, state=:state, updated_at=:updated_at`
	_, err := m.db.NamedExecContext(ctx, qry, provider)

	return err
}

func (m *ManagerDB) GetAllProviders(ctx context.Context, option *types.GetProviderOption) ([]*types.Provider, error) {
	qry := `SELECT * from providers`
	var condition []string
	if option.ID != "" {
		condition = append(condition, fmt.Sprintf(`id = '%s'`, option.ID))
	}

	if option.Owner != "" {
		condition = append(condition, fmt.Sprintf(`owner = '%s'`, option.Owner))
	}

	if len(option.State) > 0 {
		var states []string
		for _, s := range option.State {
			states = append(states, strconv.Itoa(int(s)))
		}
		condition = append(condition, fmt.Sprintf(`state in (%s)`, strings.Join(states, ",")))
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

	var out []*types.Provider
	err := m.db.SelectContext(ctx, &out, qry)
	if err != nil {
		return nil, err
	}
	return out, nil
}
