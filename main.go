package main

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/exp/event"
)

func main() {
	h := &handler{}
	exp := event.NewExporter(h, &event.ExporterOptions{
		EnableNamespaces: true,
	})
	ctx := context.Background()
	ctx = event.WithExporter(ctx, exp)

	// Log.
	event.Log(ctx, "log here", event.String("hello", "world"))

	// Counter.
	ctr := event.NewCounter("myapp_counter", &event.MetricOptions{
		Namespace:   "myapp",
		Description: "My app counter",
		Unit:        event.UnitMilliseconds,
	})
	ctr.Record(ctx, 42, event.String("GET", "/"))

	event.Annotate(ctx, event.String("annotate", "here"))

	ctx = event.Start(ctx, "start")
	ctr.Record(ctx, 3173, event.String("POST", "/"))
	event.End(ctx, event.String("end", "ends here"))
}

type handler struct {
}

func (h *handler) Event(ctx context.Context, e *event.Event) context.Context {
	fmt.Println()
	fmt.Println("###")
	fmt.Println()
	/*
		v, ok := event.MetricKey.Find(e)
		if ok {
			m, ok := v.(*event.Counter)
			if ok {
				fmt.Println("is counter")
				fmt.Println(m.Name())
				fmt.Println(m.Options().Unit)
				fmt.Println(m.Options().Description)
				fmt.Println(m.Options().Namespace)
			}
		}
	*/

	fmt.Println("> Loop start")
	for _, v := range e.Labels {
		isBasic := v.IsBool() || v.IsBytes() || v.IsDuration() || v.IsFloat64() || v.IsInt64() || v.IsString() || v.IsUint64()
		if isBasic {
			fmt.Printf("basic: %s: %v\n", v.Name, v)
		} else {
			// Must be interface type.
			c, ok := v.Interface().(*event.Counter)
			if ok {
				fmt.Println("counter", v.Name)
				fmt.Println("name:", c.Name())
				fmt.Println("options:", c.Options())
			}
		}
		fmt.Println()
	}

	fmt.Println("> Loop end")
	b, err := json.MarshalIndent(e, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", b)
	return ctx
}
