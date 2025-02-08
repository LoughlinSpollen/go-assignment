package trace

import (
	"bytes"
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"

	// "crypto/tls" // TODO configure tls for otel
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type CustomWriter struct {
	Buffer bytes.Buffer
}

func (w *CustomWriter) Write(p []byte) (n int, err error) {
	return w.Buffer.Write(p)
}

// For testing
var CustomWriterInstance *CustomWriter

func Init(ctx context.Context, otlpConfig map[string]string, serviceName, serviceVersion string) func() {
	log.Trace("Initializing distributed trace")

	var err error
	var exporter trace.SpanExporter

	if len(otlpConfig) == 0 {
		// dirty hack for testing
		if serviceName == "test" {
			CustomWriterInstance = &CustomWriter{}
			exporter, err = stdouttrace.New(
				stdouttrace.WithWriter(CustomWriterInstance),
				stdouttrace.WithPrettyPrint(),
			)
		} else {
			exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		}
		if err != nil {
			log.Fatalf("failed to initialize stdout trace: %v", err)
		}
	} else {
		// use otlp exporter
		exporter, err = otlptrace.New(
			ctx,
			otlptracehttp.NewClient(
				otlptracehttp.WithEndpoint(otlpConfig["url"]),
				// TODO: configure tls for otel in the cloud
				otlptracehttp.WithInsecure(),
			),
		)
	}
	if err != nil {
		log.Fatalf("failed to initialize otlp trace: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.TelemetrySDKLanguageGo,
		)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	)))

	return func() {
		// gracefully shut down the tracer provider
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown trace provider: %v", err)
		}
	}
}
