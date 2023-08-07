package node

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/lib/rpcenc"
	"github.com/Filecoin-Titan/titan-container/metrics"
	"github.com/Filecoin-Titan/titan-container/metrics/proxy"
	"github.com/filecoin-project/go-jsonrpc/auth"

	mhandler "github.com/Filecoin-Titan/titan-container/node/handler"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log/v2"
	"go.opencensus.io/tag"
	"golang.org/x/xerrors"
)

var rpclog = logging.Logger("rpc")

// ServeRPC serves an HTTP handler over the supplied listen multiaddr.
//
// This function spawns a goroutine to run the server, and returns immediately.
// It returns the stop function to be called to terminate the endpoint.
//
// The supplied ID is used in tracing, by inserting a tag in the context.
func ServeRPC(h http.Handler, id string, addr string) (StopFunc, error) {
	// Start listening to the addr; if invalid or occupied, we will fail early.
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, xerrors.Errorf("could not listen: %w", err)
	}

	// Instantiate the server and start listening.
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 30 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			ctx, _ := tag.New(context.Background(), tag.Upsert(metrics.APIInterface, id))
			return ctx
		},
	}

	go func() {
		err = srv.Serve(lst)
		if err != http.ErrServerClosed {
			rpclog.Warnf("rpc server failed: %s", err)
		}
	}()

	return srv.Shutdown, err
}

// ManagerHandler returns a manager handler, to be mounted as-is on the server.
func ManagerHandler(a api.Manager, permissioned bool, opts ...jsonrpc.ServerOption) (http.Handler, error) {
	m := mux.NewRouter()

	serveRpc := func(path string, hnd interface{}) {
		rpcServer := jsonrpc.NewServer(append(opts, jsonrpc.WithServerErrors(api.RPCErrors))...)
		rpcServer.Register("titan", hnd)

		var handler http.Handler = rpcServer
		if permissioned {
			handler = mhandler.New(&auth.Handler{Verify: a.AuthVerify, Next: rpcServer.ServeHTTP})
		}

		m.Handle(path, handler)
	}

	fnapi := proxy.MetricedManagerAPI(a)
	if permissioned {
		fnapi = api.PermissionedManagerAPI(fnapi)
	}

	serveRpc("/rpc/v0", fnapi)
	m.PathPrefix("/").Handler(http.DefaultServeMux) // pprof

	return m, nil
}

// ProviderHandler returns handler, to be mounted as-is on the server.
func ProviderHandler(authv func(ctx context.Context, token string) ([]auth.Permission, error), a api.Provider, permissioned bool) http.Handler {
	mux := mux.NewRouter()
	readerHandler, readerServerOpt := rpcenc.ReaderParamDecoder()
	rpcServer := jsonrpc.NewServer(jsonrpc.WithServerErrors(api.RPCErrors), readerServerOpt)

	wapi := proxy.MetricedProviderAPI(a)
	if permissioned {
		wapi = api.PermissionedProviderAPI(wapi)
	}

	rpcServer.Register("titan", wapi)
	rpcServer.AliasMethod("rpc.discover", "titan.Discover")

	mux.Handle("/rpc/v0", rpcServer)
	mux.Handle("/rpc/streams/v0/push/{uuid}", readerHandler)
	mux.PathPrefix("/").Handler(http.DefaultServeMux) // pprof

	if !permissioned {
		return mux
	}

	ah := &auth.Handler{
		Verify: authv,
		Next:   mux.ServeHTTP,
	}

	return mhandler.New(ah)
}
