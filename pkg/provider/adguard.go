package provider

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func processAdguard(filePath string) (result DataSet) {
	file, err := os.Open(filepath.Join(baseDir, filepath.Clean(filePath)))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

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

	created, err := getFileCreateTime(filePath)
	if err != nil {
		log.Fatal(err)
	}

	result.name = name
	result.data = lines
	result.count = len(lines)
	result.date = created
	result.version = date
	return result
}
