//go:build darwin || linux

package metrics

import (
	"io"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

var startTimeSeconds = float64(time.Now().UnixNano()) / 1e9

func collectUnix(w ExpfmtWriter) {
	w.WriteLazyMetricFloat64("process_start_time_seconds", startTimeSeconds)

	collectRusageUnix(w)
}

func collectRusageUnix(w ExpfmtWriter) {
	var rusage unix.Rusage

	if err := unix.Getrusage(syscall.RUSAGE_SELF, &rusage); err == nil {
		w.WriteLazyMetricDuration(
			"process_cpu_seconds_total",
			time.Duration(rusage.Stime.Nano()+rusage.Utime.Nano()),
		)
	}

	if fds, err := getOpenFileCount(); err == nil {
		w.WriteLazyMetricUint64("process_open_fds", fds)
	}

	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err == nil {
		w.WriteLazyMetricUint64("process_max_fds", rlimit.Cur)
	}
	if err := syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit); err == nil {
		w.WriteLazyMetricUint64("process_virtual_memory_max_bytes", rlimit.Cur)
	}
}

func getOpenFileCount() (uint64, error) {
	dir, err := os.Open("/dev/fd")
	if err != nil {
		return 0, err
	}
	defer dir.Close()

	var total uint64
	for {
		// Avoid ReadDir(), as it calls stat(2) on each descriptor.  Not only is
		// that info not used, but KQUEUE descriptors fail stat(2), which causes
		// the whole method to fail.
		names, err := dir.Readdirnames(512)
		total += uint64(len(names))
		switch err {
		case io.EOF:
			return total - 1, nil
		case nil:
			continue
		default:
			return 0, err
		}
	}
}
