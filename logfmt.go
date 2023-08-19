// This package demonstrates the usage of logfmt handler.
package main

import (
	"context"
	"os"

	"golang.org/x/exp/event"
	"golang.org/x/exp/event/adapter/logfmt"
	"golang.org/x/exp/event/eventtest"
)

func main() {
	opt := eventtest.ExporterOptions()

	ctx := context.Background()
	ctx = event.WithExporter(ctx, event.NewExporter(logfmt.NewHandler(os.Stdout), opt))
	event.Log(ctx, "my event", event.Int64("myInt", 6))
	event.Log(ctx, "error event", event.String("myString", "some string value"))

	c := event.NewCounter("hits", &event.MetricOptions{Description: "Earth meteorite hits"})
	c.Record(ctx, 1023)
}
