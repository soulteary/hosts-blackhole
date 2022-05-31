package provider

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/soulteary/hosts-blackhole/internal/logger"
	"github.com/soulteary/hosts-blackhole/pkg/crypto"
)

type DataSet struct {
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

func calcCacheHash(files *[]string) (string, string) {
	sort.Strings(*files)
	key := ""
	val := ""
	for _, file := range *files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			key += file
			created, _ := getFileCreateTime(file)
			val += created
		}
	}
	return crypto.Md5(key), crypto.Md5(val)
}

func updateCacheHash(files *[]string) {
	cacheKey, cacheHash = calcCacheHash(&*files)
}

func ResetCacheHash() {
	cacheKey = ""
	cacheHash = ""
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
	keys = nil
	return list
}

func Purge(files []string) (result []string, success bool) {
	log := logger.GetLogger()
	log.Info()

	if len(files) == 0 {
		return result, false
	}

	key, val := calcCacheHash(&files)
	if key == cacheKey && val == cacheHash {
		return result, false
	}

	var dataSets []DataSet
	for _, file := range files {
		types := detectType(file)
		data := DataSet{}
		switch types {
		case DATA_TYPE_STEVEN_BLACK:
			data = processStevenBlack(file)
			dataSets = append(dataSets, data)
			log.Infof("Process: %s, %s version: %s", data.name, strings.Repeat(" ", 25-len(data.name)), data.version)
			break
		case DATA_TYPE_QUIDSUP:
			data = processQuidsup(file)
			dataSets = append(dataSets, data)
			log.Infof("Process: %s, %s version: %s", data.name, strings.Repeat(" ", 25-len(data.name)), data.version)
			break
		case DATA_TYPE_ADAWAY:
			data = processAdaway(file)
			dataSets = append(dataSets, data)
			log.Infof("Process: %s, %s version: %s", data.name, strings.Repeat(" ", 25-len(data.name)), data.version)
			break
		case DATA_TYPE_ADGUARD:
			data = processAdguard(file)
			dataSets = append(dataSets, data)
			log.Infof("Process: %s, %s version: %s", data.name, strings.Repeat(" ", 25-len(data.name)), data.version)
			break
		default:
			log.Infof("Found unknown type data: %s", file)
			break
		}
	}
	updateCacheHash(&files)
	log.Info()

	log.Info("Load Providers")

	for _, data := range dataSets {
		lenKey := len(data.name)
		lenVal := len(strconv.Itoa(data.count))
		log.Infof(" - %v%s = %s %d", data.name, strings.Repeat(" ", 25-lenKey), strings.Repeat(" ", 10-lenVal), data.count)
		result = append(result, data.data...)
	}

	start := time.Now()
	total := len(result)
	result = unique(result)
	uniq := len(result)
	log.Infof("Rules uniq/total =	%d, %d", uniq, total)
	log.Info()

	elapsed := time.Since(start)
	log.Printf("data processing took %s", elapsed)
	log.Info()

	return result, true
}
