package format

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSeverityColor(t *testing.T) {
	cases := []struct {
		in   string
		want func(...any) string
	}{
		{"ERROR", red},
		{"error", red},
		{"  fatal  ", red},
		{"PANIC", red},
		{"CRITICAL", red},
		{"WARN", yellowBold},
		{"warning", yellowBold},
		{"INFO", blueBold},
		{"DEBUG", gray},
		{"TRACE", gray},
		{"unknown", whiteBold},
		{"", whiteBold},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want("X"), severityColor(tc.in)("X"))
		})
	}
}

func TestPrintPayloadHeader_AllFields(t *testing.T) {
	p := PayloadMessage{
		Body:         "oops",
		SeverityText: "ERROR",
		Timestamp:    time.Date(2026, 4, 24, 15, 41, 10, 0, time.UTC),
		Attributes:   map[string]any{"req_id": "abc", "user": "42"},
	}

	var buf bytes.Buffer
	printPayloadHeader(&buf, p)
	out := stripANSI(buf.String())

	for _, want := range []string{
		"TIMESTAMP", "2026-04-24T15:41:10",
		"SEVERITY", "ERROR",
		"BODY", "oops",
		"ATTRIBUTES",
		"req_id", "abc",
		"user", "42",
		"──",
	} {
		assert.Contains(t, out, want)
	}
}

func TestPrintPayloadHeader_OnlyBody(t *testing.T) {
	p := PayloadMessage{Body: "just body"}

	var buf bytes.Buffer
	printPayloadHeader(&buf, p)
	out := stripANSI(buf.String())

	assert.Contains(t, out, "BODY")
	assert.Contains(t, out, "just body")
	for _, missing := range []string{"TIMESTAMP", "SEVERITY", "ATTRIBUTES"} {
		assert.NotContains(t, out, missing)
	}
}

func TestPrintPayloadHeader_Empty(t *testing.T) {
	var buf bytes.Buffer
	printPayloadHeader(&buf, PayloadMessage{})
	assert.Empty(t, buf.String())
}
