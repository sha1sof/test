/*
//go:build windows
// +build windows
*/
package os

import (
	"context"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

type fileTime struct {
	lowTime  uint32
	highTime uint32
}

func GetCpuUsage(ctx context.Context) (usage uint32, currentProcessorUsage uint32, err error) {
	pid := os.Getpid()

	var idleStart, kernelStart, userStart fileTime
	err = getTotalCpuTime(&idleStart, &kernelStart, &userStart)
	if err != nil {
		return 0, 0, err
	}

	processKernelStart, processUserStart, err := getProcessCpuTime(pid)
	if err != nil {
		return 0, 0, err
	}

	select {
	case <-ctx.Done():
	}

	var idleAfter, kernelAfter, userAfter fileTime
	err = getTotalCpuTime(&idleAfter, &kernelAfter, &userAfter)
	if err != nil {
		return 0, 0, err
	}

	processKernelAfter, processUserAfter, err := getProcessCpuTime(pid)
	if err != nil {
		return 0, 0, err
	}

	totalDelta := (fileTimeToUint64(kernelAfter) + fileTimeToUint64(userAfter)) - (fileTimeToUint64(kernelStart) + fileTimeToUint64(userStart))
	idleDelta := fileTimeToUint64(idleAfter) - fileTimeToUint64(idleStart)
	processDelta := (processUserAfter + processKernelAfter) - (processUserStart - processKernelStart)

	if totalDelta > 0 {
		activeTime := totalDelta - idleDelta
		usage = uint32((float64(activeTime) / float64(totalDelta)) * 100)
	}

	if processDelta > 0 {
		currentProcessorUsage = uint32((float64(processDelta) / float64(totalDelta)) * 100)
	}

	return usage, currentProcessorUsage, nil
}

func getTotalCpuTime(idleTime, kernelTime, userTime *fileTime) error {
	result, _, err := kernel32.NewProc("GetSystemTimes").Call(
		uintptr(unsafe.Pointer(idleTime)),
		uintptr(unsafe.Pointer(kernelTime)),
		uintptr(unsafe.Pointer(userTime)))

	if result == 0 {
		return err
	}
	return nil
}

func getProcessCpuTime(pid int) (kernelTime, userTime uint64, err error) {
	header, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, 0, err
	}
	defer syscall.CloseHandle(header)

	var creation, exit, kernel, user syscall.Filetime
	err = syscall.GetProcessTimes(header, &creation, &exit, &kernel, &user)
	if err != nil {
		return 0, 0, err
	}

	return fileTimeToUint64(fileTime{kernel.LowDateTime, kernel.HighDateTime}),
		fileTimeToUint64(fileTime{user.LowDateTime, user.HighDateTime}), nil
}

func fileTimeToUint64(ft fileTime) uint64 {
	return uint64(ft.highTime)<<32 | uint64(ft.lowTime)
}

func getContextTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return time.Until(deadline)
	}
	return 1 * time.Minute
}
