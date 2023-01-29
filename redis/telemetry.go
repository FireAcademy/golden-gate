package redis

import (
    "os"
    "log"
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
    "google.golang.org/grpc/credentials"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

var (
    serviceName  = os.Getenv("SERVICE_NAME")
    collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    insecure     = os.Getenv("INSECURE_MODE")
)

func LogError(c context.Context, err error, message string) {
    span := trace.SpanFromContext(c)
    span.RecordError(err)
    span.SetStatus(codes.Error, message)
}

func GetSpan(ctx context.Context, name string) (context.Context, trace.Span) {
    return otel.Tracer(serviceName).Start(ctx, name)
}

// https://signoz.io/blog/monitoring-your-go-application-with-signoz/#instrumenting-a-sample-golang-app
func InitTracer() func(context.Context) error {
    secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
    if len(insecure) > 0 {
        secureOption = otlptracegrpc.WithInsecure()
    }

    exporter, err := otlptrace.New(
        context.Background(),
        otlptracegrpc.NewClient(
            secureOption,
            otlptracegrpc.WithEndpoint(collectorURL),
        ),
    )

    if err != nil {
        log.Fatal(err)
    }
    resources, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            attribute.String("service.name", serviceName),
            attribute.String("library.language", "go"),
        ),
    )
    if err != nil {
        log.Fatal("Could not set resources: ", err)
    }

    otel.SetTracerProvider(
        sdktrace.NewTracerProvider(
            sdktrace.WithSampler(sdktrace.AlwaysSample()),
            sdktrace.WithBatcher(exporter),
            sdktrace.WithResource(resources),
        ),
    )

    return exporter.Shutdown
}