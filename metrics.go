// Package metrics implements Prometheus-compatible metrics for applications.
//
// This package is lightweight alternative to https://github.com/prometheus/client_golang
// with simpler API and smaller dependencies.
//
// Usage:
//
//  1. Register the required metrics via New* functions.
//  2. Expose them to `/metrics` page via WritePrometheus.
//  3. Update the registered metrics during application lifetime.
//
// The package has been extracted from https://victoriametrics.com/
package metrics

import (
	"bytes"
	"cmp"
	"io"
	"slices"
	"strconv"
	"sync"
	"unsafe"
)

type namedMetric struct {
	name   string
	metric metric
	isAux  bool
}

type metric interface {
	marshalTo(prefix string, w io.Writer)
}

var (
	registeredSets     = make(map[*Set]struct{})
	registeredSetsLock sync.Mutex
)

// RegisterSet registers the given set s for metrics export via global WritePrometheus() call.
//
// See also UnregisterSet.
func RegisterSet(s *Set) {
	registeredSetsLock.Lock()
	registeredSets[s] = struct{}{}
	registeredSetsLock.Unlock()
}

// UnregisterSet stops exporting metrics for the given s via global WritePrometheus() call.
//
// If destroySet is set to true, then s.UnregisterAllMetrics() is called on s after unregistering it,
// so s becomes destroyed. Otherwise the s can be registered again in the set by passing it to RegisterSet().
func UnregisterSet(s *Set, destroySet bool) {
	registeredSetsLock.Lock()
	delete(registeredSets, s)
	registeredSetsLock.Unlock()

	if destroySet {
		s.UnregisterAllMetrics()
	}
}

// WritePrometheus writes all the metrics in Prometheus format from the default set, all the added sets and metrics writers to w.
//
// Additional sets can be registered via RegisterSet() call.
// Additional metric writers can be registered via RegisterMetricsWriter() call.
//
// If exposeProcessMetrics is true, then various `go_*` and `process_*` metrics
// are exposed for the current process.
//
// The WritePrometheus func is usually called inside "/metrics" handler:
//
//	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
//	    metrics.WritePrometheus(w, true)
//	})
func WritePrometheus(w io.Writer, exposeProcessMetrics bool) {
	registeredSetsLock.Lock()
	sets := make([]*Set, 0, len(registeredSets))
	for s := range registeredSets {
		sets = append(sets, s)
	}
	registeredSetsLock.Unlock()

	slices.SortFunc(sets, func(a, b *Set) int {
		return cmp.Compare(uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(b)))
	})

	for _, s := range sets {
		s.WritePrometheus(w)
	}
	if exposeProcessMetrics {
		WriteProcessMetrics(w)
	}
}

// WriteProcessMetrics writes additional process metrics in Prometheus format to w.
//
// The following `go_*` and `process_*` metrics are exposed for the currently
// running process. Below is a short description for the exposed `process_*` metrics:
//
//   - process_cpu_seconds_system_total - CPU time spent in syscalls
//
//   - process_cpu_seconds_user_total - CPU time spent in userspace
//
//   - process_cpu_seconds_total - CPU time spent by the process
//
//   - process_major_pagefaults_total - page faults resulted in disk IO
//
//   - process_minor_pagefaults_total - page faults resolved without disk IO
//
//   - process_resident_memory_bytes - recently accessed memory (aka RSS or resident memory)
//
//   - process_resident_memory_peak_bytes - the maximum RSS memory usage
//
//   - process_resident_memory_anon_bytes - RSS for memory-mapped files
//
//   - process_resident_memory_file_bytes - RSS for memory allocated by the process
//
//   - process_resident_memory_shared_bytes - RSS for memory shared between multiple processes
//
//   - process_virtual_memory_bytes - virtual memory usage
//
//   - process_virtual_memory_peak_bytes - the maximum virtual memory usage
//
//   - process_num_threads - the number of threads
//
//   - process_start_time_seconds - process start time as unix timestamp
//
//   - process_io_read_bytes_total - the number of bytes read via syscalls
//
//   - process_io_written_bytes_total - the number of bytes written via syscalls
//
//   - process_io_read_syscalls_total - the number of read syscalls
//
//   - process_io_write_syscalls_total - the number of write syscalls
//
//   - process_io_storage_read_bytes_total - the number of bytes actually read from disk
//
//   - process_io_storage_written_bytes_total - the number of bytes actually written to disk
//
//   - go_sched_latencies_seconds - time spent by goroutines in ready state before they start execution
//
//   - go_mutex_wait_seconds_total - summary time spent by all the goroutines while waiting for locked mutex
//
//   - go_gc_mark_assist_cpu_seconds_total - summary CPU time spent by goroutines in GC mark assist state
//
//   - go_gc_cpu_seconds_total - summary time spent in GC
//
//   - go_gc_pauses_seconds - duration of GC pauses
//
//   - go_scavenge_cpu_seconds_total - CPU time spent on returning the memory to OS
//
//   - go_memlimit_bytes - the GOMEMLIMIT env var value
//
//   - go_memstats_alloc_bytes - memory usage for Go objects in the heap
//
//   - go_memstats_alloc_bytes_total - the cumulative counter for total size of allocated Go objects
//
//   - go_memstats_buck_hash_sys_bytes - bytes of memory in profiling bucket hash tables
//
//   - go_memstats_frees_total - the cumulative counter for number of freed Go objects
//
//   - go_memstats_gc_cpu_fraction - the fraction of CPU spent in Go garbage collector
//
//   - go_memstats_gc_sys_bytes - the size of Go garbage collector metadata
//
//   - go_memstats_heap_alloc_bytes - the same as go_memstats_alloc_bytes
//
//   - go_memstats_heap_idle_bytes - idle memory ready for new Go object allocations
//
//   - go_memstats_heap_inuse_bytes - bytes in in-use spans
//
//   - go_memstats_heap_objects - the number of Go objects in the heap
//
//   - go_memstats_heap_released_bytes - bytes of physical memory returned to the OS
//
//   - go_memstats_heap_sys_bytes - memory requested for Go objects from the OS
//
//   - go_memstats_last_gc_time_seconds - unix timestamp the last garbage collection finished
//
//   - go_memstats_lookups_total - the number of pointer lookups performed by the runtime
//
//   - go_memstats_mallocs_total - the number of allocations for Go objects
//
//   - go_memstats_mcache_inuse_bytes - bytes of allocated mcache structures
//
//   - go_memstats_mcache_sys_bytes - bytes of memory obtained from the OS for mcache structures
//
//   - go_memstats_mspan_inuse_bytes - bytes of allocated mspan structures
//
//   - go_memstats_mspan_sys_bytes - bytes of memory obtained from the OS for mspan structures
//
//   - go_memstats_next_gc_bytes - the target heap size when the next garbage collection should start
//
//   - go_memstats_other_sys_bytes - bytes of memory in miscellaneous off-heap runtime allocations
//
//   - go_memstats_stack_inuse_bytes - memory used for goroutine stacks
//
//   - go_memstats_stack_sys_bytes - memory requested fromthe OS for goroutine stacks
//
//   - go_memstats_sys_bytes - memory requested by Go runtime from the OS
//
//   - go_cgo_calls_count - the total number of CGO calls
//
//   - go_cpu_count - the number of CPU cores on the host where the app runs
//
// The WriteProcessMetrics func is usually called in combination with writing Set metrics
// inside "/metrics" handler:
//
//	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
//	    mySet.WritePrometheus(w)
//	    metrics.WriteProcessMetrics(w)
//	})
//
// See also WriteFDMetrics.
func WriteProcessMetrics(w io.Writer) {
	writeGoMetrics(w)
	writeProcessMetrics(w)
}

// WriteFDMetrics writes `process_max_fds` and `process_open_fds` metrics to w.
func WriteFDMetrics(w io.Writer) {
	writeFDMetrics(w)
}

func isFloatInteger(v float64) bool {
	return float64(int64(v)) == v
}

// WriteMetricUint64 writes metric with the given name and value to w in Prometheus text exposition format.
func WriteMetricUint64(w io.Writer, metricName string, value uint64) {
	bb, isBuffer := w.(*bytes.Buffer)
	if !isBuffer {
		bb = getMetricsBuffer()
		defer freeMetricsBuffer(bb)
	}

	bb.WriteString(metricName)
	bb.WriteByte(' ')
	buf := bb.AvailableBuffer()
	buf = strconv.AppendUint(buf, value, 10)
	bb.Write(buf)
	bb.WriteByte('\n')

	if !isBuffer {
		w.Write(bb.Bytes())
	}
}

// WriteMetricInt64 writes metric with the given name and value to w in Prometheus text exposition format.
func WriteMetricInt64(w io.Writer, metricName string, value int64) {
	bb, isBuffer := w.(*bytes.Buffer)
	if !isBuffer {
		bb = getMetricsBuffer()
		defer freeMetricsBuffer(bb)
	}

	bb.WriteString(metricName)
	bb.WriteByte(' ')
	buf := bb.AvailableBuffer()
	buf = strconv.AppendInt(buf, value, 10)
	bb.Write(buf)
	bb.WriteByte('\n')

	if !isBuffer {
		w.Write(bb.Bytes())
	}
}

// WriteMetricFloat64 writes metric with the given name and value to w in Prometheus text exposition format.
func WriteMetricFloat64(w io.Writer, metricName string, value float64) {
	bb, isBuffer := w.(*bytes.Buffer)
	if !isBuffer {
		bb = getMetricsBuffer()
		defer freeMetricsBuffer(bb)
	}

	bb.WriteString(metricName)
	bb.WriteByte(' ')
	buf := bb.AvailableBuffer()
	buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
	bb.Write(buf)
	bb.WriteByte('\n')

	if !isBuffer {
		w.Write(bb.Bytes())
	}
}

var metricsBufferPool sync.Pool

const defaultBufferSize = 256

func getMetricsBuffer() *bytes.Buffer {
	if bb := metricsBufferPool.Get(); bb != nil {
		return bb.(*bytes.Buffer)
	}
	return bytes.NewBuffer(make([]byte, defaultBufferSize))
}

func freeMetricsBuffer(bb *bytes.Buffer) {
	bb.Reset()
	metricsBufferPool.Put(bb)
}
