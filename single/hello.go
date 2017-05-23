package main

import (
	"fmt"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
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
	parent := opentracing.GlobalTracer().StartSpan("hello")
	defer parent.Finish()
	childFunction(parent)
}

func childFunction(parent opentracing.Span) {
	child := opentracing.GlobalTracer().StartSpan(
		"world", opentracing.ChildOf(parent.Context()))
	time.Sleep(1 * time.Second)
	defer child.Finish()
}
