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
	DATA_TYPE_UNKNOWN      int = 0
	DATA_TYPE_STEVEN_BLACK     = 1
	DATA_TYPE_QUIDSUP          = 2
	DATA_TYPE_ADAWAY           = 3
	DATA_TYPE_ADGUARD          = 4
)

const baseDir = "./"

var cacheKey = ""
var cacheHash = ""

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

func detectType(filePath string) int {
	file, err := os.Open(filepath.Join(baseDir, filepath.Clean(filePath)))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	lineNumber := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if lineNumber > 30 {
			return DATA_TYPE_UNKNOWN
		}
		if strings.HasPrefix(scanner.Text(), "# Title: StevenBlack/hosts") {
			return DATA_TYPE_STEVEN_BLACK
		}
		if strings.HasPrefix(scanner.Text(), "# Title: NoTrack") {
			return DATA_TYPE_QUIDSUP
		}
		if strings.HasPrefix(scanner.Text(), "# AdAway default blocklist") {
			return DATA_TYPE_ADAWAY
		}
		if strings.HasPrefix(scanner.Text(), "! Title: AdGuard DNS filter") {
			return DATA_TYPE_ADGUARD
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return DATA_TYPE_UNKNOWN
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

func ResetCacheHash() {
	cacheKey = ""
	cacheHash = ""
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
		case DATA_TYPE_STEVEN_BLACK:
			result = caseStevenBlack(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case DATA_TYPE_QUIDSUP:
			result = caseQuidsup(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case DATA_TYPE_ADAWAY:
			result = caseAdaway(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		case DATA_TYPE_ADGUARD:
			result = caseAdguard(file)
			results = append(results, result)
			log.Infof("Process: %s, %s version: %s", result.name, strings.Repeat(" ", 25-len(result.name)), result.version)
			break
		default:
			log.Infof("Found unknown type data: %s", file)
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
