package modules

import (
	"github.com/Filecoin-Titan/titan-container/db"
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	logging "github.com/ipfs/go-log/v2"
	"github.com/jmoiron/sqlx"
)

var log = logging.Logger("modules")

// NewManagerDB creates a new database connection for managing managers.
// It takes a DSN (Data Source Name) string as input and returns a pointer to sqlx.DB and an error.
func NewManagerDB(dsn string) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		return db.SqlDB(dsn)
	}
}

// NewSetManagerConfigFunc creates a function to set the manager config
func NewSetManagerConfigFunc(r repo.LockedRepo) func(cfg config.ManagerCfg) error {
	return func(cfg config.ManagerCfg) (err error) {
		return r.SetConfig(func(raw interface{}) {
			_, ok := raw.(*config.ManagerCfg)
			if !ok {
				return
			}
		})
	}
}

// NewGetManagerConfigFunc creates a function to get the manager config
func NewGetManagerConfigFunc(r repo.LockedRepo) func() (config.ManagerCfg, error) {
	return func() (out config.ManagerCfg, err error) {
		raw, err := r.Config()
		if err != nil {
			return
		}

		scfg, ok := raw.(*config.ManagerCfg)
		if !ok {
			return
		}

		out = *scfg
		return
	}
}
