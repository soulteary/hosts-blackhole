package provider

import (
	"bufio"
	"crypto/md5" //#nosec
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/soulteary/hosts-blackhole/internal/logger"
)

type Lines struct {
	name    string
	data    []string
	count   int
	date    string
	version string
}

const (
	StevenBlack string = "steven-black"
	Quidsup            = "quidsup"
	Adaway             = "adaway"
	Adguard            = "adguard"
)

const baseDir = "./"

var cacheKey = ""
var cacheHash = ""

func getCreateTime(filePath string) (string, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(filePath, &st); err != nil {
		return "", err
	}

	return time.Unix(st.Ctimespec.Sec, 0).Format("02/01/2006, 15:04:05"), nil
}

func calcMd5(str string) string {
	data := []byte(str)
	/* #nosec */
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func CacheHash(files []string, update bool) (string, string) {
	sort.Strings(files)
	key := ""
	val := ""
	for _, file := range files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			key += file
			created, _ := getCreateTime(file)
			val += created
		}
	}

	key = calcMd5(key)
	val = calcMd5(val)

	if update {
		cacheKey = key
		cacheHash = val
	}
	return key, val
}

func detectType(filePath string) string {
	file, err := os.Open(filepath.Join(baseDir, filepath.Clean(filePath)))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lineNumber := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if lineNumber > 30 {
			return ""
		}
		if strings.HasPrefix(scanner.Text(), "# Title: StevenBlack/hosts") {
			return StevenBlack
		}
		if strings.HasPrefix(scanner.Text(), "# Title: NoTrack") {
			return Quidsup
		}
		if strings.HasPrefix(scanner.Text(), "# AdAway default blocklist") {
			return Adaway
		}
		if strings.HasPrefix(scanner.Text(), "! Title: AdGuard DNS filter") {
			return Adguard
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}

func unique(src []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range src {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Purge(files []string) (mixed []string, success bool) {
	log := logger.GetLogger()
	log.Info()

	if len(files) == 0 {
		return mixed, false
	}

	key, val := CacheHash(files, false)
	if key == cacheKey && val == cacheHash {
		return mixed, false
	}

	var results []Lines
	for _, file := range files {
		types := detectType(file)
		result := Lines{}
		switch types {
		case StevenBlack:
			result = caseStevenBlack(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case Quidsup:
			result = caseQuidsup(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case Adaway:
			result = caseAdaway(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case Adguard:
			result = caseAdguard(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		}
	}
	CacheHash(files, true)
	log.Info()

	log.Info("Load Providers")

	for _, result := range results {
		lenKey := len(result.name)
		lenVal := len(strconv.Itoa(result.count))
		log.Infof(" - %v%s = %s %d", result.name, strings.Repeat(" ", 25-lenKey), strings.Repeat(" ", 10-lenVal), result.count)
		mixed = append(mixed, result.data...)
	}

	start := time.Now()
	total := len(mixed)
	mixed = unique(mixed)
	unique := len(mixed)
	log.Infof("Rules uniq/total =	%d, %d", unique, total)
	log.Info()

	elapsed := time.Since(start)
	log.Printf("data processing took %s", elapsed)
	log.Info()

	return mixed, true
}

func ManualGC() {
	log := logger.GetLogger()

	log.Info("Runtime Information:")

	runtime.GC()
	debug.FreeOSMemory()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Infof(" MEM Alloc =        %10v MB", toMB(m.Alloc))
	log.Infof(" MEM HeapAlloc =    %10v MB", toMB(m.HeapAlloc))
	log.Infof(" MEM Sys =          %10v MB", toMB(m.Sys))
	log.Infof(" MEM NumGC =        %10v", m.NumGC)
	log.Infof(" RUN NumCPU =       %10d", runtime.NumCPU())
	log.Infof(" RUN NumGoroutine = %10d", runtime.NumGoroutine())
}

func toMB(b uint64) uint64 {
	const bytesInKB = 1024
	return b / bytesInKB / bytesInKB
}
