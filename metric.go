// This example demonstrates the metric provider handler.
package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric/metrictest"
	"golang.org/x/exp/event"
	"golang.org/x/exp/event/otel"
)

func main() {
	mp := metrictest.NewMeterProvider()
	mh := otel.NewMetricHandler(mp.Meter("test"))

	ctx := context.Background()
	ctx = event.WithExporter(ctx, event.NewExporter(mh, nil))
	event.Log(ctx, "my event", event.Int64("myInt", 6))
	event.Log(ctx, "error event", event.String("myString", "some string value"))

	c := event.NewCounter("hits", &event.MetricOptions{Description: "Earth meteorite hits"})
	c.Record(ctx, 1023)

	got := metrictest.AsStructs(mp.MeasurementBatches)
	fmt.Printf("%#v", got)
}
