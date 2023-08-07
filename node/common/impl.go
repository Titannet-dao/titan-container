package common

import (
	"context"
	"fmt"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/Filecoin-Titan/titan-container/build"
	"github.com/Filecoin-Titan/titan-container/journal/alerting"
	"github.com/Filecoin-Titan/titan-container/node/modules/dtypes"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gbrlsnchs/jwt/v3"
	"go.uber.org/fx"

	"github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
)

var session = uuid.New()

// CommonAPI api o
type CommonAPI struct {
	fx.In

	Alerting     *alerting.Alerting
	APISecret    *dtypes.APIAlg
	ShutdownChan dtypes.ShutdownChan
}

// MethodGroup: Auth

type jwtPayload struct {
	Allow []auth.Permission
}

// AuthVerify verifies a JWT token and returns the permissions associated with it
func (a *CommonAPI) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	var payload jwtPayload
	if _, err := jwt.Verify([]byte(token), (*jwt.HMACSHA)(a.APISecret), &payload); err != nil {
		return nil, xerrors.Errorf("JWT Verification failed: %w", err)
	}

	return payload.Allow, nil
}

// AuthNew generates a new JWT token with the provided permissions
func (a *CommonAPI) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	p := jwtPayload{
		Allow: perms, // TODO: consider checking validity
	}

	return jwt.Sign(&p, (*jwt.HMACSHA)(a.APISecret))
}

// LogList returns a list of available logging subsystems
func (a *CommonAPI) LogList(context.Context) ([]string, error) {
	return logging.GetSubsystems(), nil
}

// LogSetLevel sets the log level for a given subsystem
func (a *CommonAPI) LogSetLevel(ctx context.Context, subsystem, level string) error {
	return logging.SetLogLevel(subsystem, level)
}

// LogAlerts returns an empty list of alerts
func (a *CommonAPI) LogAlerts(ctx context.Context) ([]alerting.Alert, error) {
	return []alerting.Alert{}, nil
}

// Version provides information about API provider
func (a *CommonAPI) Version(context.Context) (api.APIVersion, error) {
	v, err := api.VersionForType(types.RunningNodeType)
	if err != nil {
		return api.APIVersion{}, err
	}

	return api.APIVersion{
		Version:    build.UserVersion(),
		APIVersion: v,
	}, nil
}

// Discover returns an OpenRPC document describing an RPC API.
func (a *CommonAPI) Discover(ctx context.Context) (types.OpenRPCDocument, error) {
	return nil, fmt.Errorf("not implement")
}

// Shutdown trigger graceful shutdown
func (a *CommonAPI) Shutdown(context.Context) error {
	a.ShutdownChan <- struct{}{}
	return nil
}

// Session returns a UUID of api provider session
func (a *CommonAPI) Session(ctx context.Context) (uuid.UUID, error) {
	return session, nil
}

// Closing jsonrpc closing
func (a *CommonAPI) Closing(context.Context) (<-chan struct{}, error) {
	return make(chan struct{}), nil // relies on jsonrpc closing
}
