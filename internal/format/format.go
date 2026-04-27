package format

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/fatih/color"
)

var (
	cyan       = color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow     = color.New(color.FgYellow).SprintFunc()
	yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	green      = color.New(color.FgGreen).SprintFunc()
	gray       = color.New(color.FgHiBlack).SprintFunc()
	whiteBold  = color.New(color.FgWhite, color.Bold).SprintFunc()
	magenta    = color.New(color.FgMagenta).SprintFunc()
	red        = color.New(color.FgRed, color.Bold).SprintFunc()
	blueBold   = color.New(color.FgBlue, color.Bold).SprintFunc()
)

func Print(w io.Writer, data []byte) {
	var (
		stackTraceData string
		payloadMsg     PayloadMessage
	)

	if json.Valid(data) {
		if err := json.Unmarshal(data, &payloadMsg); err != nil {
			log.Fatal(err)
		}
		stackTraceAttr, ok := payloadMsg.Attributes["StackTrace"]

		if !ok {
			log.Fatal(errors.New("StackTrace not in Attributes"))
		}

		stackTraceData = stackTraceAttr.(string)

		printPayloadHeader(w, payloadMsg)
	} else {
		stackTraceData = string(data)
	}

	var (
		pendingFunc        string
		pendingIsCreatedBy bool
		needBlank          bool
	)

	emit := func(funcLine, fileLine string, isCreatedBy bool) {
		if needBlank {
			fmt.Fprintln(w)
		}
		isUser := fileLine != "" &&
			!strings.Contains(fileLine, "/usr/local/go/src/") &&
			!strings.Contains(fileLine, "/go/pkg/mod/")

		switch {
		case isCreatedBy:
			fmt.Fprintln(w, magenta(funcLine))
		case isUser:
			fmt.Fprintln(w, yellowBold(funcLine))
		default:
			fmt.Fprintln(w, yellow(funcLine))
		}
		if fileLine != "" {
			fmt.Fprintln(w, formatFileLine(fileLine, isUser))
		}
		needBlank = true
	}

	flushPending := func() {
		if pendingFunc != "" {
			emit(pendingFunc, "", pendingIsCreatedBy)
			pendingFunc = ""
			pendingIsCreatedBy = false
		}
	}

	stackTraceData = strings.ReplaceAll(stackTraceData, "\\n", "\n")

	for line := range strings.SplitSeq(stackTraceData, "\n") {
		line = strings.ReplaceAll(line, "\\t", "\t")
		if strings.TrimSpace(line) == "" {
			continue
		}

		switch {
		case isGoroutineHeader(line):
			flushPending()
			if needBlank {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, cyan(line))
			needBlank = false

		case strings.HasPrefix(line, "\t"):
			if pendingFunc != "" {
				emit(pendingFunc, line, pendingIsCreatedBy)
				pendingFunc = ""
				pendingIsCreatedBy = false
			} else {
				if needBlank {
					fmt.Fprintln(w)
				}
				fmt.Fprintln(w, gray(line))
				needBlank = true
			}

		default:
			flushPending()
			pendingFunc = line
			pendingIsCreatedBy = strings.HasPrefix(line, "created by ")
		}
	}
	flushPending()
}

func isGoroutineHeader(s string) bool {
	return strings.HasPrefix(s, "goroutine ") && strings.HasSuffix(s, ":")
}

func formatFileLine(line string, isUser bool) string {
	rest := strings.TrimPrefix(line, "\t")
	var suffix string
	if i := strings.LastIndex(rest, " "); i > 0 {
		suffix = rest[i:]
		rest = rest[:i]
	}
	var dir, fileColon string
	if i := strings.LastIndex(rest, "/"); i >= 0 {
		dir = rest[:i+1]
		fileColon = rest[i+1:]
	} else {
		fileColon = rest
	}
	dirColored := gray(dir)
	if isUser {
		dirColored = green(dir)
	}
	return "\t" + dirColored + whiteBold(fileColon) + gray(suffix)
}
