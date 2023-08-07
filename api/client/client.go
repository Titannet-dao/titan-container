package client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/Filecoin-Titan/titan-container/api"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/Filecoin-Titan/titan-container/lib/rpcenc"
)

// NewManager creates a new http jsonrpc client.
func NewManager(ctx context.Context, addr string, requestHeader http.Header) (api.Manager, jsonrpc.ClientCloser, error) {
	pushURL, err := getPushURL(addr)
	if err != nil {
		return nil, nil, err
	}

	var res api.ManagerStruct
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "titan",
		api.GetInternalStructs(&res),
		requestHeader,
		rpcenc.ReaderParamEncoder(pushURL),
	)

	return &res, closer, err
}

func getPushURL(addr string) (string, error) {
	pushURL, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	switch pushURL.Scheme {
	case "ws":
		pushURL.Scheme = "http"
	case "wss":
		pushURL.Scheme = "https"
	}
	///rpc/v0 -> /rpc/streams/v0/push

	pushURL.Path = path.Join(pushURL.Path, "../streams/v0/push")
	return pushURL.String(), nil
}

// NewProvider creates a new http jsonrpc client for provider
func NewProvider(ctx context.Context, addr string, requestHeader http.Header, opts ...jsonrpc.Option) (api.Provider, jsonrpc.ClientCloser, error) {
	pushURL, err := getPushURL(addr)
	if err != nil {
		return nil, nil, err
	}

	var res api.ProviderStruct
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "titan",
		api.GetInternalStructs(&res),
		requestHeader,
		rpcenc.ReaderParamEncoder(pushURL),
		jsonrpc.WithTimeout(30*time.Second),
		jsonrpc.WithErrors(api.RPCErrors),
	)

	return &res, closer, err
}

// NewCommonRPCV0 creates a new http jsonrpc client.
func NewCommonRPCV0(ctx context.Context, addr string, requestHeader http.Header) (api.Common, jsonrpc.ClientCloser, error) {
	var res api.CommonStruct
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "titan",
		api.GetInternalStructs(&res),
		requestHeader)

	return &res, closer, err
}
