package node

import (
	"errors"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/Filecoin-Titan/titan-container/node/impl/provider"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	"go.uber.org/fx"

	"golang.org/x/xerrors"
)

func Provider(out *api.Provider) Option {
	return Options(
		ApplyIf(func(s *Settings) bool { return s.Config },
			Error(errors.New("the Provider option must be set before Config option")),
		),

		func(s *Settings) error {
			s.nodeType = repo.Provider
			return nil
		},

		func(s *Settings) error {
			resAPI := &provider.Provider{}
			s.invokes[ExtractAPIKey] = fx.Populate(resAPI)
			*out = resAPI
			return nil
		},
	)
}

func ConfigProvider(c interface{}) Option {
	cfg, ok := c.(*config.ProviderCfg)
	if !ok {
		return Error(xerrors.Errorf("invalid config from repo, got: %T", c))
	}

	return Options(
		ConfigCommon(&cfg.Common),
		Override(new(*config.ProviderCfg), cfg),
		Override(new(provider.Manager), provider.NewManager),
	)
}
