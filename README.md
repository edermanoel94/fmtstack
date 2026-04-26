# fmtstack

A tiny Go CLI that reads a Go stack trace from your clipboard and pretty-prints it with ANSI colors so **user-code frames stand out** from runtime and dependency frames.

## Install

```bash
go install github.com/edermanoel94/fmtstack@latest
```

Or from source:

```bash
git clone https://github.com/edermanoel94/fmtstack
cd fmtstack
go build -o fmtstack .
```

> On Linux the clipboard library needs X11/Wayland headers at build time (`apt install libx11-dev` on Debian/Ubuntu).

## Usage

1. Copy a Go stack trace to your clipboard. Both shapes work:
   - **Raw** — the output of `runtime/debug.Stack()` or a panic dump.
   - **JSON envelope** — a log line with `Body`, `SeverityText`, `Stacktrace`, `Attributes`, `time` fields (e.g. copied from an observability tool).
2. Run:

```bash
fmtstack
```

Or pipe the trace in instead of using the clipboard:

```bash
cat panic.log | fmtstack --stdin
kubectl logs my-pod | fmtstack --stdin
```

For JSON payloads, a header (timestamp, severity, body, attributes) is printed above the trace.

## How it works

`main.go` is the whole program:

1. Read clipboard bytes.
2. If they're valid JSON, unmarshal and pull the `Stacktrace` field; otherwise treat the bytes as the trace.
3. Walk the trace line by line, classifying each as a goroutine header, function line, or `file:line` line.
4. Color each frame. A frame is **user code** if its file path contains **neither** `/usr/local/go/src/` **nor** `/go/pkg/mod/`.

That's it. If you want to understand what `{_, _}` or `+0x2c` mean in the trace itself, see [`STACKTRACE.md`](STACKTRACE.md).
