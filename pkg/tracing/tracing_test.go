package tracing

import (
	"context"
	"reflect"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TestSerializeTraceCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want TraceEnvelope
	}{
		{
			name: "should serialize empty trace context",
			args: args{
				ctx: context.Background(),
			},
			want: TraceEnvelope{},
		},
		{
			name: "should serialize trace context with traceparent",
			args: args{
				ctx: func() context.Context {
					otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

					carrier := propagation.MapCarrier{}
					carrier.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
					return otel.GetTextMapPropagator().Extract(context.Background(), carrier)
				}(),
			},
			want: TraceEnvelope{
				TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SerializeTraceCtx(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SerializeTraceCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeserializeTraceCtx(t *testing.T) {
	type args struct {
		env TraceEnvelope
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "should deserialize empty trace envelope",
			args: args{
				env: TraceEnvelope{},
			},
			want: context.Background(),
		},
		{
			name: "should deserialize trace envelope with traceparent",
			args: args{
				env: TraceEnvelope{
					TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				},
			},
			want: func() context.Context {
				otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

				carrier := propagation.MapCarrier{}
				carrier.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
				return otel.GetTextMapPropagator().Extract(context.Background(), carrier)
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeserializeTraceCtx(tt.args.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeTraceCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
