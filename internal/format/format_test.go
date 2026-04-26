package format

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

func TestMain(m *testing.M) {
	color.NoColor = false
	os.Exit(m.Run())
}

func TestIsGoroutineHeader(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"goroutine 1 [running]:", true},
		{"goroutine 67 [chan receive, 5 minutes]:", true},
		{"goroutine 1 [running]", false},
		{"main.foo()", false},
		{"", false},
		{"goroutine", false},
		{"goroutine 1 [running]:extra", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, isGoroutineHeader(tc.in))
		})
	}
}

func TestFormatFileLine_TextContent(t *testing.T) {
	cases := []struct {
		name   string
		line   string
		isUser bool
		want   string
	}{
		{
			name:   "user code with offset",
			line:   "\t/home/eder/x/foo.go:10 +0x2c",
			isUser: true,
			want:   "\t/home/eder/x/foo.go:10 +0x2c",
		},
		{
			name:   "stdlib without offset",
			line:   "\t/usr/local/go/src/runtime/panic.go:787",
			isUser: false,
			want:   "\t/usr/local/go/src/runtime/panic.go:787",
		},
		{
			name:   "no directory",
			line:   "\tno_dir.go:5",
			isUser: true,
			want:   "\tno_dir.go:5",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stripANSI(formatFileLine(tc.line, tc.isUser)))
		})
	}
}

func TestFormatFileLine_UserDirIsGreen(t *testing.T) {
	out := formatFileLine("\t/home/eder/x/foo.go:10 +0x2c", true)
	assert.Contains(t, out, green("/home/eder/x/"), "user directory should be green")
}

func TestFormatFileLine_DepDirIsGray(t *testing.T) {
	out := formatFileLine("\t/usr/local/go/src/runtime/panic.go:787", false)
	assert.Contains(t, out, gray("/usr/local/go/src/runtime/"), "dep directory should be gray")
}

func TestPrint_Raw(t *testing.T) {
	in := []byte("goroutine 1 [running]:\n" +
		"main.foo()\n" +
		"\t/home/eder/x/main.go:10 +0x2c\n" +
		"main.main()\n" +
		"\t/home/eder/x/main.go:5 +0x18\n")

	var buf bytes.Buffer
	Print(&buf, in)
	out := stripANSI(buf.String())

	for _, want := range []string{
		"goroutine 1 [running]:",
		"main.foo()",
		"/home/eder/x/main.go:10 +0x2c",
		"main.main()",
		"/home/eder/x/main.go:5 +0x18",
	} {
		assert.Contains(t, out, want)
	}
}

func TestPrint_JSONHeaderAndTrace(t *testing.T) {
	payload := `{"time":"2026-04-24T15:41:10.619311-03:00",` +
		`"Body":"oops","SeverityText":"ERROR",` +
		`"Stacktrace":"goroutine 1 [running]:\nmain.foo()\n\t/home/x/main.go:10 +0x2c\n",` +
		`"Attributes":{"req_id":"abc","user":"42"}}`

	var buf bytes.Buffer
	Print(&buf, []byte(payload))
	out := stripANSI(buf.String())

	for _, want := range []string{
		"TIMESTAMP", "2026-04-24",
		"SEVERITY", "ERROR",
		"BODY", "oops",
		"ATTRIBUTES", "req_id", "abc", "user", "42",
		"goroutine 1 [running]:",
		"main.foo()",
	} {
		assert.Contains(t, out, want)
	}

	iReq := strings.Index(out, "req_id")
	iUser := strings.Index(out, "user")
	assert.True(t, iReq >= 0 && iReq < iUser,
		"attributes should be sorted alphabetically: req_id at %d, user at %d", iReq, iUser)
}

func TestPrint_EscapedNewlines(t *testing.T) {
	payload := `{"Stacktrace":"goroutine 1 [running]:\\nmain.foo()\\n\\t/home/x/main.go:10 +0x2c\\n"}`

	var buf bytes.Buffer
	Print(&buf, []byte(payload))
	out := stripANSI(buf.String())

	for _, want := range []string{
		"goroutine 1 [running]:",
		"main.foo()",
		"\t/home/x/main.go:10 +0x2c",
	} {
		assert.Contains(t, out, want)
	}
}

func TestPrint_CreatedByGetsMagenta(t *testing.T) {
	in := []byte("goroutine 1 [running]:\n" +
		"created by main.foo in goroutine 8\n" +
		"\t/home/x/main.go:10 +0x73\n")

	var buf bytes.Buffer
	Print(&buf, in)

	assert.Contains(t, buf.String(), magenta("created by main.foo in goroutine 8"),
		"'created by' line should be magenta")
}
