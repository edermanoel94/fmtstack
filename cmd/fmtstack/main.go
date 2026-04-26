package main

import (
	"io"
	"log"
	"os"

	"github.com/edermanoel94/fmtstack/internal/format"
	"golang.design/x/clipboard"
)

func main() {
	var data []byte
	if len(os.Args) > 1 && os.Args[1] == "--stdin" {
		var err error
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err := clipboard.Init(); err != nil {
			log.Fatal(err)
		}
		data = clipboard.Read(clipboard.FmtText)
	}

	format.Print(os.Stdout, data)
}
