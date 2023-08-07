package node

import (
	"errors"

	"github.com/Filecoin-Titan/titan-container/db"
	"github.com/Filecoin-Titan/titan-container/node/impl/manager"
	"github.com/Filecoin-Titan/titan-container/node/modules"
	"github.com/Filecoin-Titan/titan-container/node/modules/dtypes"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
)

func Manager(out *api.Manager) Option {
	return Options(
		ApplyIf(func(s *Settings) bool { return s.Config },
			Error(errors.New("the Manager option must be set before Config option")),
		),

		func(s *Settings) error {
			s.nodeType = repo.Manager
			return nil
		},

		func(s *Settings) error {
			resAPI := &manager.Manager{}
			s.invokes[ExtractAPIKey] = fx.Populate(resAPI)
			*out = resAPI
			return nil
		},
	)
}

func ConfigManager(c interface{}) Option {
	cfg, ok := c.(*config.ManagerCfg)
	if !ok {
		return Error(xerrors.Errorf("invalid config from repo, got: %T", c))
	}

	return Options(
		Override(new(*config.ManagerCfg), cfg),
		ConfigCommon(&cfg.Common),
		Override(new(*sqlx.DB), modules.NewManagerDB(cfg.DatabaseAddress)),
		Override(new(*db.ManagerDB), db.NewManagerDB),
		Override(new(*manager.ProviderManager), manager.NewProviderScheduler),
		Override(new(dtypes.SetManagerConfigFunc), modules.NewSetManagerConfigFunc),
		Override(new(dtypes.GetManagerConfigFunc), modules.NewGetManagerConfigFunc),
	)
}
