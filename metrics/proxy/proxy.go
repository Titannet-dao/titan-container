package proxy

import (
	"context"
	"reflect"

	"github.com/Filecoin-Titan/titan-container/api"
	"github.com/Filecoin-Titan/titan-container/metrics"

	"go.opencensus.io/tag"
)

func MetricedManagerAPI(a api.Manager) api.Manager {
	var out api.ManagerStruct
	proxy(a, &out)
	return &out
}

func MetricedProviderAPI(a api.Provider) api.Provider {
	var out api.ProviderStruct
	proxy(a, &out)
	return &out
}

func proxy(in interface{}, outstr interface{}) {
	outs := api.GetInternalStructs(outstr)
	for _, out := range outs {
		rint := reflect.ValueOf(out).Elem()
		ra := reflect.ValueOf(in)

		for f := 0; f < rint.NumField(); f++ {
			field := rint.Type().Field(f)
			fn := ra.MethodByName(field.Name)

			rint.Field(f).Set(reflect.MakeFunc(field.Type, func(args []reflect.Value) (results []reflect.Value) {
				ctx := args[0].Interface().(context.Context)
				// upsert function name into context
				ctx, _ = tag.New(ctx, tag.Upsert(metrics.Endpoint, field.Name))
				stop := metrics.Timer(ctx, metrics.APIRequestDuration)
				defer stop()
				// pass tagged ctx back into function call
				args[0] = reflect.ValueOf(ctx)
				return fn.Call(args)
			}))
		}
	}
}
