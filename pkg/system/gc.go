package system

import (
	"runtime"
	"runtime/debug"

	"github.com/soulteary/hosts-blackhole/internal/logger"
)

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
