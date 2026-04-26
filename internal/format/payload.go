package format

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type PayloadMessage struct {
	Body         string            `json:"Body"`
	SeverityText string            `json:"SeverityText"`
	StackTrace   string            `json:"Stacktrace"`
	Attributes   map[string]string `json:"Attributes"`
	Timestamp    time.Time         `json:"time"`
}

func severityColor(s string) func(a ...any) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "ERROR", "FATAL", "PANIC", "CRITICAL":
		return red
	case "WARN", "WARNING":
		return yellowBold
	case "INFO":
		return blueBold
	case "DEBUG", "TRACE":
		return gray
	default:
		return whiteBold
	}
}

func printPayloadHeader(w io.Writer, p PayloadMessage) {
	label := func(s string) string {
		return gray(fmt.Sprintf("%-10s", s))
	}

	printed := false
	if !p.Timestamp.IsZero() {
		fmt.Fprintln(w, label("TIMESTAMP"), whiteBold(p.Timestamp.Format(time.RFC3339Nano)))
		printed = true
	}
	if p.SeverityText != "" {
		fmt.Fprintln(w, label("SEVERITY"), severityColor(p.SeverityText)(p.SeverityText))
		printed = true
	}
	if p.Body != "" {
		fmt.Fprintln(w, label("BODY"), whiteBold(p.Body))
		printed = true
	}
	if len(p.Attributes) > 0 {
		fmt.Fprintln(w, label("ATTRIBUTES"))
		keys := make([]string, 0, len(p.Attributes))
		maxKey := 0
		for k := range p.Attributes {
			keys = append(keys, k)
			if len(k) > maxKey {
				maxKey = len(k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "  %s %s %s\n",
				yellow(fmt.Sprintf("%-*s", maxKey, k)),
				gray("="),
				whiteBold(p.Attributes[k]))
		}
		printed = true
	}
	if printed {
		fmt.Fprintln(w, gray(strings.Repeat("─", 60)))
		fmt.Fprintln(w)
	}
}
