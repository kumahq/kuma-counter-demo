package main

import (
	"context"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	"github.com/kumahq/kuma-counter-demo/app/internal/base"
	"github.com/kumahq/kuma-counter-demo/app/public"
	"github.com/kumahq/kuma-counter-demo/pkg/api"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	version    = ""
	kvUrl      = ""
	appAddress = ""
	appPort    = "5050"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}
	otel.SetLogger(logr.FromSlogHandler(slog.Default().Handler()))
	// Set up propagator.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3.New(),
	))

	tracerProvider, err := newTraceProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}

func newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	// Set up trace provider.
	traceClient := otlptracegrpc.NewClient(otlptracegrpc.WithInsecure())

	traceExporter, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	prometheusExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	//otlpExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure(), otlpmetrichttp.WithCompression(otlpmetrichttp.NoCompression))
	//if err != nil {
	//	return nil, err
	//}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(prometheusExporter),
		//metric.WithReader(metric.NewPeriodicReader(otlpExporter)),
	)
	return meterProvider, nil
}

// Middleware to introduce delay based on the header value
func delayMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delayStr := r.Header.Get("x-set-response-delay-ms")
		delay, _ := strconv.Atoi(delayStr)
		if delay > 0 {
			slog.DebugContext(r.Context(), "simulating slow response", "time", delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))
	if v := os.Getenv("APP_VERSION"); v != "" {
		version = v
	}
	if c := os.Getenv("ADDRESS"); c != "" {
		appAddress = c
	}
	if p := os.Getenv("PORT"); p != "" {
		appPort = p
	}
	if h := os.Getenv("KV_URL"); h != "" {
		kvUrl = h
	}

	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	r := mux.NewRouter()

	swagger, err := api.GetSwagger()
	if err != nil {
		return err
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = []*openapi3.Server{
		{
			URL: "/api",
		},
	}
	http.DefaultClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Create an instance of our handler which satisfies the generated interface
	apiHandler := base.NewServerImpl(slog.Default(), kvUrl, version)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(otelmux.Middleware("api-server"))
	// Serve static files from the "public" directory
	// Use our validation middleware to check all requests against the OpenAPI schema.
	apiRouter.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{SilenceServersWarning: true}))
	apiRouter.Use(mux.CORSMethodMiddleware(r))
	apiRouter.Use(delayMiddleware)
	api.HandlerFromMux(apiHandler, apiRouter)
	r.PathPrefix("/metrics").Handler(promhttp.Handler())
	r.PathPrefix("/").Methods("GET").Handler(http.FileServer(http.FS(public.Files)))

	// Apply middlewares

	addr := net.JoinHostPort(appAddress, appPort)
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Graceful shutdown handling
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()
	slog.Info("server running", "addr", addr)
	select {
	case err = <-srvErr:
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}
	err = srv.Shutdown(context.Background())
	return
}
