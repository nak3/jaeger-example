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

type data struct {
	tracer opentracing.Tracer
	ctx    context.Context
}

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
		"test_example",
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	ctx := context.TODO()
	d := &data{
		tracer: tracer,
		ctx:    ctx,
	}
	span := d.tracer.StartSpan("Init")
	defer span.Finish()
	d.ctx = opentracing.ContextWithSpan(ctx, span)
	if span := opentracing.SpanFromContext(d.ctx); span != nil {
		span := d.tracer.StartSpan("hello", opentracing.ChildOf(span.Context()))
		span.SetTag("first func", "hello")
		defer span.Finish()
		d.ctx = opentracing.ContextWithSpan(d.ctx, span)
	} else {
		// NG
		fmt.Println("ng1")
	}
	d.childFunction()
	time.Sleep(1 * time.Second)
}

func (d *data) childFunction() {
	if span := opentracing.SpanFromContext(d.ctx); span != nil {
		span := d.tracer.StartSpan("world", opentracing.ChildOf(span.Context()))
		span.SetTag("second func", "test")
		defer span.Finish()
		d.ctx = opentracing.ContextWithSpan(d.ctx, span)
		span.LogFields(
			log.String("event", "soft error"),
			log.String("type", "cache timeout"),
			log.Int("waited.millis", 1500))
	} else {
		// NG
		fmt.Println("ng2")
	}
	time.Sleep(1 * time.Second)
}
