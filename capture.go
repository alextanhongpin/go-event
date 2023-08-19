package main

import (
	"fmt"

	"golang.org/x/exp/event"
	"golang.org/x/exp/event/eventtest"
)

func main() {
	ctx, ec := eventtest.NewCapture()
	event.Log(ctx, "my event", event.Int64("myInt", 6))
	event.Log(ctx, "error event", event.String("myString", "some string value"))

	c := event.NewCounter("hits", &event.MetricOptions{Description: "Earth meteorite hits"})
	c.Record(ctx, 1023)

	fmt.Printf("%#v", ec.Got)
}
