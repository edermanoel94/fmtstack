package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

const (
	chunkSize = 2048
)

func main() {

	bufReader := bufio.NewReader(os.Stdin)

	buf := make([]byte, chunkSize)

	_, err := bufReader.Read(buf)

	if err != nil {
		log.Fatal(err)
	}

	data := string(buf)

	lines := strings.Split(data, "\\n")

	for _, line := range lines {

		if len(line) == 0 {
			continue
		}

		cleaned := strings.Replace(line, "\\t", "	", 1)

		color.Red(cleaned)

	}
}
