package provider

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

func caseAdaway(filePath string) (result Lines) {
	file, err := os.Open(filePath)
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

	var st syscall.Stat_t
	if err := syscall.Stat(filePath, &st); err != nil {
		log.Fatal(err)
	}

	result.name = name
	result.date = time.Unix(st.Ctimespec.Sec, 0).Format("02/01/2006, 15:04:05")
	result.data = lines
	result.count = len(lines)
	return result
}
