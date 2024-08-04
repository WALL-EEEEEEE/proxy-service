package metric

import (
	"context"

	metric "go.opentelemetry.io/otel/metric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var meters *sdk_metric.MeterProvider
var stats *sdk_metric.ManualReader

func newMeterProvider() *sdk_metric.MeterProvider {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("spider-services-proxy"),
			semconv.ServiceVersion("0.1.0"),
		))
	if err != nil {
		panic(err)
	}
	stats = sdk_metric.NewManualReader()
	/* exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	reader := sdk_metric.WithReader(sdk_metric.NewPeriodicReader(exporter,
		// Default is 1m. Set to 3s for demonstrative purposes.
		sdk_metric.WithInterval(3*time.Second)))
	*/
	meters = sdk_metric.NewMeterProvider(
		sdk_metric.WithResource(res),
		sdk_metric.WithReader(stats),
	)
	return meters
}

func getMeterProvider() *sdk_metric.MeterProvider {
	if meters != nil {
		return meters
	}
	meters = newMeterProvider()
	return meters

}

func NewMeter(name string) metric.Meter {
	meters := getMeterProvider()
	meter := meters.Meter(name)
	return meter
}

func Collect(ctx context.Context, callback func(*metricdata.ResourceMetrics)) {
	resource_metrics := &metricdata.ResourceMetrics{}
	err := stats.Collect(ctx, resource_metrics)
	if err != nil {
		panic(err)
	}
	callback(resource_metrics)
}
