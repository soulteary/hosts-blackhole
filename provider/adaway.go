package provider

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func caseAdaway(filePath string) (result Lines) {
	file, err := os.Open(filepath.Join(baseDir, filepath.Clean(filePath)))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	skipLocalRecords := false
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	name := "AdAway default blocklist"

	for scanner.Scan() {
		line := scanner.Text()

		if !skipLocalRecords && strings.HasPrefix(line, "::1  localhost") {
			skipLocalRecords = true
			continue
		}

		if skipLocalRecords {
			if !strings.HasPrefix(line, "#") {
				if parts := strings.Fields(line); len(parts) == 2 {
					host := parts[1]
					ip := net.ParseIP(host)
					if ip == nil {
						lines = append(lines, strings.TrimSpace(strings.ToLower(host)))
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	created, err := getCreateTime(filePath)
	if err != nil {
		log.Fatal(err)
	}

	result.name = name
	result.data = lines
	result.count = len(lines)
	result.date = created
	result.version = result.date
	return result
}
