/*
//go:build linux
// +build linux
*/
package os

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetCpuUsageL(ctx context.Context) (usage, currentProcessorUsage uint32, err error) {
	pid := os.Getpid()

	idleStart, totalStart, err := getTotalCpuTimeL()
	if err != nil {
		return 0, 0, err
	}

	processStart, err := getProcessCpuTimeL(pid)
	if err != nil {
		return 0, 0, err
	}

	select {
	case <-time.After(getContextTimeout(ctx)):
	}

	idleEnd, totalEnd, err := getTotalCpuTimeL()
	if err != nil {
		return 0, 0, err
	}

	processEnd, err := getProcessCpuTimeL(pid)
	if err != nil {
		return 0, 0, err
	}

	totalDelta := totalEnd - totalStart
	idleDelta := idleEnd - idleStart
	processDelta := processEnd - processStart

	if totalDelta > 0 {
		activeTime := totalDelta - idleDelta
		usage = uint32((float64(activeTime) / float64(totalDelta)) * 100)
	}

	if processDelta > 0 {
		currentProcessorUsage = uint32((float64(processDelta) / float64(totalDelta)) * 100)
	}

	return usage, currentProcessorUsage, nil
}

func getTotalCpuTimeL() (idle uint32, total uint32, err error) {
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0, 0, fmt.Errorf("unexpected format of /proc/stat")
			}

			var totalTime uint64
			for _, field := range fields[1:] {
				value, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return 0, 0, err
				}
				totalTime += value
			}

			idleTime, err := strconv.ParseUint(fields[4], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			return uint32(idleTime), uint32(totalTime), nil
		}
	}

	return 0, 0, fmt.Errorf("cpu line not found in /proc/stat")
}

func getProcessCpuTimeL(pid int) (uint32, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))

	utime, err := strconv.ParseUint(fields[13], 10, 64)
	if err != nil {
		return 0, err
	}

	stime, err := strconv.ParseUint(fields[14], 10, 64)
	if err != nil {
		return 0, err
	}

	return uint32(utime + stime), nil
}

func getContextTimeoutL(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return time.Until(deadline)
	}
	return 1 * time.Minute
}
