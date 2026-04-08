package agent

import (
	"math/rand/v2"
	"runtime"
	"sync"
)

// MetricsCollector collects runtime metrics from the Go runtime.
//
// generate:reset
type MetricsCollector struct {
	mu          sync.RWMutex
	gauges      map[string]float64
	pollCount   int64
	randomValue float64
	rand        *rand.Rand
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		gauges: make(map[string]float64),
		rand:   rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func (mc *MetricsCollector) Collect() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mc.gauges[MetricAlloc] = float64(memStats.Alloc)
	mc.gauges[MetricBuckHashSys] = float64(memStats.BuckHashSys)
	mc.gauges[MetricFrees] = float64(memStats.Frees)
	mc.gauges[MetricGCCPUFraction] = memStats.GCCPUFraction
	mc.gauges[MetricGCSys] = float64(memStats.GCSys)
	mc.gauges[MetricHeapAlloc] = float64(memStats.HeapAlloc)
	mc.gauges[MetricHeapIdle] = float64(memStats.HeapIdle)
	mc.gauges[MetricHeapInuse] = float64(memStats.HeapInuse)
	mc.gauges[MetricHeapObjects] = float64(memStats.HeapObjects)
	mc.gauges[MetricHeapReleased] = float64(memStats.HeapReleased)
	mc.gauges[MetricHeapSys] = float64(memStats.HeapSys)
	mc.gauges[MetricLastGC] = float64(memStats.LastGC)
	mc.gauges[MetricLookups] = float64(memStats.Lookups)
	mc.gauges[MetricMCacheInuse] = float64(memStats.MCacheInuse)
	mc.gauges[MetricMCacheSys] = float64(memStats.MCacheSys)
	mc.gauges[MetricMSpanInuse] = float64(memStats.MSpanInuse)
	mc.gauges[MetricMSpanSys] = float64(memStats.MSpanSys)
	mc.gauges[MetricMallocs] = float64(memStats.Mallocs)
	mc.gauges[MetricNextGC] = float64(memStats.NextGC)
	mc.gauges[MetricNumForcedGC] = float64(memStats.NumForcedGC)
	mc.gauges[MetricNumGC] = float64(memStats.NumGC)
	mc.gauges[MetricOtherSys] = float64(memStats.OtherSys)
	mc.gauges[MetricPauseTotalNs] = float64(memStats.PauseTotalNs)
	mc.gauges[MetricStackInuse] = float64(memStats.StackInuse)
	mc.gauges[MetricStackSys] = float64(memStats.StackSys)
	mc.gauges[MetricSys] = float64(memStats.Sys)
	mc.gauges[MetricTotalAlloc] = float64(memStats.TotalAlloc)

	mc.randomValue = mc.rand.Float64()
	mc.pollCount++
}

func (mc *MetricsCollector) GetGauges() map[string]float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]float64, len(mc.gauges)+1)
	for k, v := range mc.gauges {
		result[k] = v
	}
	result[MetricRandomValue] = mc.randomValue
	return result
}

func (mc *MetricsCollector) GetPollCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.pollCount
}
