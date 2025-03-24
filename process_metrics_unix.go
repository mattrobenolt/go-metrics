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

func collectUnix(w ExpfmtWriter, constantTags string) {
	w.WriteMetricFloat64(MetricName{
		Family:       MustIdent("process_start_time_seconds"),
		ConstantTags: constantTags,
	}, startTimeSeconds)

	collectRusageUnix(w, constantTags)
}

func collectRusageUnix(w ExpfmtWriter, constantTags string) {
	var rusage unix.Rusage

	if err := unix.Getrusage(syscall.RUSAGE_SELF, &rusage); err == nil {
		w.WriteMetricDuration(MetricName{
			Family:       MustIdent("process_cpu_seconds_total"),
			ConstantTags: constantTags,
		}, time.Duration(rusage.Stime.Nano()+rusage.Utime.Nano()))
	}

	if fds, err := getOpenFileCount(); err == nil {
		w.WriteMetricUint64(MetricName{
			Family:       MustIdent("process_open_fds"),
			ConstantTags: constantTags,
		}, fds)
	}

	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err == nil {
		w.WriteMetricUint64(MetricName{
			Family:       MustIdent("process_max_fds"),
			ConstantTags: constantTags,
		}, rlimit.Cur)
	}
	if err := syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit); err == nil {
		w.WriteMetricUint64(MetricName{
			Family:       MustIdent("process_virtual_memory_max_bytes"),
			ConstantTags: constantTags,
		}, rlimit.Cur)
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
