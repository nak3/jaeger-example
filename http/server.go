package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

type data struct {
	tracer opentracing.Tracer
}

// TracedServeMux is a wrapper around http.ServeMux that instruments handlers for tracing.
type TracedServeMux struct {
	mux    *http.ServeMux
	tracer opentracing.Tracer
}

// Handle implements http.ServeMux#Handle
func (tm *TracedServeMux) Handle(pattern string, handler http.Handler) {
	middleware := nethttp.Middleware(
		tm.tracer,
		handler,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + " " + pattern
		}))
	tm.mux.Handle(pattern, middleware)
}

// ServeHTTP implements http.ServeMux#ServeHTTP
func (tm *TracedServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm.mux.ServeHTTP(w, r)
}

// NewServeMux creates a new TracedServeMux.
func NewServeMux(tracer opentracing.Tracer) *TracedServeMux {
	return &TracedServeMux{
		mux:    http.NewServeMux(),
		tracer: tracer,
	}
}

func main() {
	d := &data{tracer: Init()}
	mux := NewServeMux(d.tracer)
	mux.Handle("/", http.HandlerFunc(d.root))
	fmt.Println("Starting httpd server ... port 8888")
	http.ListenAndServe(":8888", http.Handler(mux))
}

func (d *data) root(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	d.firstFunction(ctx)
	d.secondFunction(ctx)
}

func Init() opentracing.Tracer {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	tracer, _, err := cfg.New(
		"hello_world_service",
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return tracer
}

func (d *data) firstFunction(ctx context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := d.tracer.StartSpan("hello", opentracing.ChildOf(span.Context()))
		span.SetTag("first func", "hello")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	time.Sleep(1 * time.Second)
}

func (d *data) secondFunction(ctx context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := d.tracer.StartSpan("world", opentracing.ChildOf(span.Context()))
		span.SetTag("second func", "test")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	time.Sleep(1 * time.Second)
}
