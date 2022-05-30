package provider

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

func caseAdguard(filePath string) (result Lines) {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	getUpdateTime := false
	date := ""
	getFileName := false
	name := ""
	scanner := bufio.NewScanner(file)

	lines := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()

		if !getFileName && strings.HasPrefix(line, "! Title:") {
			name = line[len("! Title: "):]
			getFileName = true
			continue
		}

		if strings.HasPrefix(line, "!") {
			if !getUpdateTime && strings.HasPrefix(line, "! Last modified:") {
				date = line[len("! Last modified: "):]
				getUpdateTime = true
			}
			continue
		}

		if strings.HasPrefix(line, "||") && strings.HasSuffix(line, "^") {
			host := line[2 : len(line)-1]
			ip := net.ParseIP(host)
			if ip == nil {
				lines = append(lines, strings.TrimSpace(strings.ToLower(host)))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	result.name = name
	result.date = date
	result.data = lines
	result.count = len(lines)
	return result
}
