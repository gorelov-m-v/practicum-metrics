package agent

import (
	"fmt"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type GopsutilCollector struct {
	mu     sync.RWMutex
	gauges map[string]float64
	cpuNum int
}

func NewGopsutilCollector() *GopsutilCollector {
	cpuCount, _ := cpu.Counts(true)
	return &GopsutilCollector{
		gauges: make(map[string]float64),
		cpuNum: cpuCount,
	}
}

func (gc *GopsutilCollector) Collect() error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory stats: %w", err)
	}

	gc.gauges[MetricTotalMemory] = float64(vmStat.Total)
	gc.gauges[MetricFreeMemory] = float64(vmStat.Free)

	percentages, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("failed to get CPU percentages: %w", err)
	}

	for i, percent := range percentages {
		metricName := fmt.Sprintf("%s%d", MetricCPUutilization1, i+1)
		gc.gauges[metricName] = percent
	}

	return nil
}

func (gc *GopsutilCollector) GetGauges() map[string]float64 {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	result := make(map[string]float64, len(gc.gauges))
	for k, v := range gc.gauges {
		result[k] = v
	}
	return result
}
