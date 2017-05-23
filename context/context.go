package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			QueueSize:           10,
		},
	}

	tracer, closer, err := cfg.New(
		"first_example",
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	ctx := context.Background()
	parent := opentracing.GlobalTracer().StartSpan("hello")
	defer parent.Finish()
	childFunction(parent)
	childFunction(parent)
	childFunction(parent)
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := tracer.StartSpan("hello", opentracing.ChildOf(span.Context()))
		span.SetTag("first func", "hello")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	time.Sleep(1 * time.Second)
}

func childFunction(parent opentracing.Span) {
	span := opentracing.GlobalTracer().StartSpan(
		"world", opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	span.LogFields(
		log.String("event", "soft error"),
		log.String("type", "cache timeout"),
		log.Int("waited.millis", 1500))
	span.LogFields(
		log.String("event2", "hard error"),
		log.String("type2", "cache timeout"),
		log.Int("waited.millis", 100))
	time.Sleep(1 * time.Second)
}
