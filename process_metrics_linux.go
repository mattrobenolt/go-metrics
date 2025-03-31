//go:build linux

package metrics

import (
	"bytes"
	"fmt"
	"os"
)

// Different environments may have different page size.
//
// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/6457
var pageSizeBytes = uint64(os.Getpagesize())

// See http://man7.org/linux/man-pages/man5/proc.5.html
type procStat struct {
	State       byte
	Ppid        int
	Pgrp        int
	Session     int
	TtyNr       int
	Tpgid       int
	Flags       uint
	Minflt      uint
	Cminflt     uint
	Majflt      uint
	Cmajflt     uint
	Utime       uint
	Stime       uint
	Cutime      int
	Cstime      int
	Priority    int
	Nice        int
	NumThreads  int
	ItrealValue int
	Starttime   uint64
	Vsize       uint
	Rss         int
}

func (c *processMetricsCollector) Collect(w ExpfmtWriter) {
	collectUnix(w)
	collectStatMetrics(w)
}

func collectStatMetrics(w ExpfmtWriter) {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return
	}

	// Search for the end of command.
	n := bytes.LastIndex(data, []byte(") "))
	if n < 0 {
		return
	}
	data = data[n+2:]

	var p procStat
	bb := bytes.NewBuffer(data)
	_, err = fmt.Fscanf(
		bb,
		"%c %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
		&p.State,
		&p.Ppid,
		&p.Pgrp,
		&p.Session,
		&p.TtyNr,
		&p.Tpgid,
		&p.Flags,
		&p.Minflt,
		&p.Cminflt,
		&p.Majflt,
		&p.Cmajflt,
		&p.Utime,
		&p.Stime,
		&p.Cutime,
		&p.Cstime,
		&p.Priority,
		&p.Nice,
		&p.NumThreads,
		&p.ItrealValue,
		&p.Starttime,
		&p.Vsize,
		&p.Rss,
	)
	if err != nil {
		return
	}

	w.WriteLazyMetricUint64("process_resident_memory_bytes", uint64(p.Rss)*pageSizeBytes)
	w.WriteLazyMetricUint64("process_virtual_memory_bytes", uint64(p.Vsize))
}
