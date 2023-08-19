package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"log/slog"

	"go.opentelemetry.io/otel/metric/metrictest"
	"golang.org/x/exp/event"
	"golang.org/x/exp/event/adapter/logfmt"
	"golang.org/x/exp/event/eventtest"
	"golang.org/x/exp/event/otel"
)

func main() {
	mp := metrictest.NewMeterProvider()
	mh := otel.NewMetricHandler(mp.Meter("test"))

	log := logfmt.NewHandler(os.Stdout)

	opt := eventtest.ExporterOptions()
	opt.EnableNamespaces = true

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	ctx := context.Background()
	//_ = log
	ctx = event.WithExporter(ctx, event.NewExporter(&multiHandler{
		metric: mh,
		log:    log,
		slog:   NewSlogHandler(logger),
	}, opt))
	event.Log(ctx, "my event", event.Int64("myInt", 6))
	event.Log(ctx, "error event", event.String("myString", "some string value"))
	event.Error(ctx, "hello", errors.New("unexpected error has occured"))

	c := event.NewCounter("hits", &event.MetricOptions{Description: "Earth meteorite hits"})
	c.Record(ctx, 1023)

	got := metrictest.AsStructs(mp.MeasurementBatches)
	fmt.Printf("%#v", got)
}

type multiHandler struct {
	metric *otel.MetricHandler
	trace  *otel.TraceHandler
	slog   *SlogHandler
	log    *logfmt.Handler
}

func (h *multiHandler) Event(ctx context.Context, ev *event.Event) context.Context {
	if h.slog != nil {
		ctx = h.slog.Event(ctx, ev)
	}

	if h.metric != nil {
		ctx = h.metric.Event(ctx, ev)
	}

	if h.trace != nil {
		ctx = h.trace.Event(ctx, ev)
	}

	if h.log != nil {
		ctx = h.log.Event(ctx, ev)
	}

	return ctx
}

type SlogHandler struct {
	logger *slog.Logger
}

func NewSlogHandler(logger *slog.Logger) *SlogHandler {
	return &SlogHandler{
		logger: logger,
	}
}

func (h *SlogHandler) Event(ctx context.Context, ev *event.Event) context.Context {
	if ev.Kind != event.LogKind {
		return ctx
	}

	var attrs []slog.Attr

	if ev.Source.Space != "" {
		attrs = append(attrs, slog.String("in", ev.Source.Space))
	}

	if ev.Source.Owner != "" {
		attrs = append(attrs, slog.String("owner", ev.Source.Owner))
	}

	if ev.Source.Name != "" {
		attrs = append(attrs, slog.String("name", ev.Source.Name))
	}
	if ev.Parent != 0 {
		attrs = append(attrs, slog.Uint64("parent", ev.Parent))
	}

	var isError bool
	for _, l := range ev.Labels {
		if !l.HasValue() || l.Name == "" || l.Name == "msg" {
			continue
		}

		if l.Name == "error" {
			isError = true
		}

		attrs = append(attrs, label(l))
	}

	level := slog.LevelInfo
	if isError {
		level = slog.LevelError
	}

	msg := ev.Find("msg").String()
	h.logAttrs(ctx, ev.At, level, msg, attrs...)

	return ctx
}

func (h *SlogHandler) logAttrs(ctx context.Context, at time.Time, level slog.Level, msg string, attrs ...slog.Attr) {
	l := h.logger
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	// skip [runtime.Callers, this function logAttrs, this function's caller Event]
	runtime.Callers(7, pcs[:])
	pc := pcs[0]
	r := slog.NewRecord(at, level, msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

func label(l event.Label) slog.Attr {
	if !l.HasValue() || l.Name == "" {
		return slog.Attr{}
	}

	switch {
	case l.IsString():
		return slog.String(l.Name, l.String())
	case l.IsBytes():
		return slog.String(l.Name, string(l.Bytes()))
	case l.IsInt64():
		return slog.Int64(l.Name, l.Int64())
	case l.IsUint64():
		return slog.Uint64(l.Name, l.Uint64())
	case l.IsFloat64():
		return slog.Float64(l.Name, l.Float64())
	case l.IsBool():
		return slog.Bool(l.Name, l.Bool())
	default:
		v := l.Interface()
		switch v := v.(type) {
		case string:
			return slog.String(l.Name, v)
		case fmt.Stringer:
			return slog.String(l.Name, v.String())
		default:
			return slog.Any(l.Name, v)
		}
	}
}
