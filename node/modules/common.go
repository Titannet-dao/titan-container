package modules

import (
	"context"

	"github.com/Filecoin-Titan/titan-container/build"
	"github.com/Filecoin-Titan/titan-container/lib/ulimit"
	"github.com/Filecoin-Titan/titan-container/node/repo"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
)

const (
	// ServerIDName Server ID key name in the keystore
	ServerIDName = "server-id" //nolint:gosec
	// KTServerIDSecret Key type for server ID secret
	KTServerIDSecret = "server-id-secret" //nolint:gosec
)

// LockedRepo returns a function that returns the locked repository with an added lifecycle hook to close the repository
func LockedRepo(lr repo.LockedRepo) func(lc fx.Lifecycle) repo.LockedRepo {
	return func(lc fx.Lifecycle) repo.LockedRepo {
		lc.Append(fx.Hook{
			OnStop: func(_ context.Context) error {
				return lr.Close()
			},
		})

		return lr
	}
}

// CheckFdLimit checks the file descriptor limit and returns an error if the limit is too low
func CheckFdLimit() error {
	limit, _, err := ulimit.GetLimit()
	switch {
	case err == ulimit.ErrUnsupported:
		log.Errorw("checking file descriptor limit failed", "error", err)
	case err != nil:
		return xerrors.Errorf("checking fd limit: %w", err)
	default:
		if limit < build.EdgeFDLimit {
			return xerrors.Errorf("soft file descriptor limit (ulimit -n) too low, want %d, current %d", build.EdgeFDLimit, limit)
		}
	}
	return nil
}
