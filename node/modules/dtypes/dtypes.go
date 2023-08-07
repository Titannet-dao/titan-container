package dtypes

import (
	"github.com/Filecoin-Titan/titan-container/node/config"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multiaddr"
)

// MetadataDS stores metadata.
type MetadataDS datastore.Batching

type APIAlg jwt.HMACSHA

type APIEndpoint multiaddr.Multiaddr

type ProviderID string

// InternalIP local network address
type InternalIP string

type (
	NodeMetadataPath string
	AssetsPaths      []string
)

// SetManagerConfigFunc is a function which is used to
// sets the manager config.
type SetManagerConfigFunc func(cfg config.ManagerCfg) error

// GetManagerConfigFunc is a function which is used to
// get the sealing config.
type GetManagerConfigFunc func() (config.ManagerCfg, error)
