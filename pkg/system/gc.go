package system

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/soulteary/hosts-blackhole/internal/logger"
)

func Stats(logPrint bool) string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := []string{
		"Runtime Information:",
		fmt.Sprintf(" MEM Alloc =        %10v MB", toMB(m.Alloc)),
		fmt.Sprintf(" MEM HeapAlloc =    %10v MB", toMB(m.HeapAlloc)),
		fmt.Sprintf(" MEM Sys =          %10v MB", toMB(m.Sys)),
		fmt.Sprintf(" MEM NumGC =        %10v", m.NumGC),
		fmt.Sprintf(" RUN NumCPU =       %10d", runtime.NumCPU()),
		fmt.Sprintf(" RUN NumGoroutine = %10d", runtime.NumGoroutine()),
	}

	if logPrint {
		log := logger.GetLogger()
		for _, info := range stats {
			log.Info(info)
		}
	}
	return strings.Join(stats, "\n")
}

func ManualGC() {
	runtime.GC()
	debug.FreeOSMemory()
	Stats(true)
}

func toMB(b uint64) uint64 {
	const bytesInKB = 1024
	return b / bytesInKB / bytesInKB
}
