package prettylog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"
	"log/slog"

	"github.com/fatih/color"
)

type prettyHandler struct {
	opts   *slog.HandlerOptions
	out    io.Writer
	groups []string
	attrs  []slog.Attr
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *prettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &prettyHandler{
		out:  out,
		opts: opts,
	}
}

func (h *prettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *prettyHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// Format time
	timestamp := r.Time.Format("["+time.DateTime+"]")

	if _, err := buf.WriteString(color.WhiteString(timestamp + " ")); err != nil {
		return err
	}

	// Format level
	levelStr := formatLevel(r.Level)
	if _, err := buf.WriteString(levelStr + " "); err != nil {
		return err
	}

	// Format message
	if _, err := buf.WriteString(color.WhiteString(r.Message)); err != nil {
		return err
	}

	// Format source
	if h.opts.AddSource {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			src := fmt.Sprintf(" (%s:%d)", f.File, f.Line)
			if _, err := buf.WriteString(color.MagentaString(src)); err != nil {
				return err
			}
		}
	}

	// Collect attributes
	attrs := h.collectAttrs(r)

	// Format attributes
	if len(attrs) > 0 {
		buf.WriteString("\n")
		h.formatAttributes(&buf, attrs, 1)
	}

	buf.WriteString("\n")

	_, err := h.out.Write(buf.Bytes())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{
		opts:   h.opts,
		out:    h.out,
		groups: h.groups,
		attrs:  append(h.attrs, attrs...),
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{
		opts:   h.opts,
		out:    h.out,
		groups: append(h.groups, name),
		attrs:  h.attrs,
	}
}

func (h *prettyHandler) collectAttrs(r slog.Record) map[string]any {
	attrs := make(map[string]any)

	// Add handler attributes
	for _, attr := range h.attrs {
		h.addAttr(attrs, attr)
	}

	// Add record attributes
	r.Attrs(func(attr slog.Attr) bool {
		h.addAttr(attrs, attr)
		return true
	})

	return attrs
}

func (h *prettyHandler) addAttr(attrs map[string]any, attr slog.Attr) {
	key := attr.Key
	if len(h.groups) > 0 {
		key = strings.Join(h.groups, ".") + "." + key
	}

	if attr.Value.Kind() == slog.KindGroup {
		for _, groupAttr := range attr.Value.Group() {
			h.addAttr(attrs, groupAttr)
		}
	} else {
		attrs[key] = attr.Value.Any()
	}
}

func (h *prettyHandler) formatAttributes(buf *bytes.Buffer, attrs map[string]any, indent int) {
	for key, value := range attrs {
		if nested, ok := value.(map[string]any); ok {
			buf.WriteString(strings.Repeat("  ", indent))
			buf.WriteString(color.CyanString("%s:", key) + "\n")
			h.formatAttributes(buf, nested, indent+1)
		} else {
			buf.WriteString(strings.Repeat("  ", indent))
			buf.WriteString(fmt.Sprintf("%s: %+v\n",
				color.CyanString(key),
				value,
			))
		}
	}
}

func formatLevel(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return color.BlueString("DEBUG")
	case level < slog.LevelWarn:
		return color.GreenString("INFO")
	case level < slog.LevelError:
		return color.YellowString("WARN")
	default:
		return color.RedString("ERROR")
	}
}
