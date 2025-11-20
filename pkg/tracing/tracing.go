package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TraceEnvelope struct {
	TraceParent string `json:"traceparent"`
	TraceState  string `json:"tracestate,omitempty"`
}

func SerializeTraceCtx(ctx context.Context) TraceEnvelope {
	carrier := propagation.MapCarrier{}

	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return TraceEnvelope{
		TraceParent: carrier["traceparent"],
	}
}

func DeserializeTraceCtx(env TraceEnvelope) context.Context {
	carrier := propagation.MapCarrier{
		"traceparent": env.TraceParent,
	}

	ctx := context.Background()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	return ctx
}
