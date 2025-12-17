package promhttp

import (
	"bufio"
	"io"
	"slices"
	"strings"
	"sync"
)

// Type is a metric type to be described
type Type uint8

const (
	Untyped Type = iota
	Counter
	Gauge
	Histogram
	Summary
	Info
)

func (t Type) String() string {
	switch t {
	case Untyped:
		return "untyped"
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	case Summary:
		return "summary"
	case Info:
		return "info"
	default:
		panic("unknown Type")
	}
}

// Desc is a description of a metric family
type Desc struct {
	Help string
	Type Type
}

// Mapping is a map of metric family names to their descriptions
type Mapping map[string]Desc

func (m Mapping) get(family string) (string, Desc) {
	// check if we have an exact match in our mapping first
	if d, ok := m[family]; ok {
		return family, d
	}

	family = normalizeFamily(family)
	return family, m[family]
}

// A Transformer accepts Prometheus metrics written in, and can read out
// metrics augmented by a [Mapping]. If a mapping does not exist for a metric
// family, it is assumed to be [Untyped].
//
// A zero value Transformer is usable with no [Mapping].
type Transformer struct {
	pr           io.ReadCloser
	pw           io.WriteCloser
	mapping      Mapping
	initR, initW sync.Once
}

type eofReadWriter struct{}

func (eofReadWriter) Read(p []byte) (n int, err error)  { return 0, io.EOF }
func (eofReadWriter) Write(p []byte) (n int, err error) { return 0, io.EOF }
func (eofReadWriter) Close() error                      { return nil }

// Read must not be called concurrently with Write. Read also must not
// be called before writing has completed.
func (t *Transformer) Read(p []byte) (n int, err error) {
	t.initR.Do(func() {
		if t.pw == nil {
			t.pr = eofReadWriter{}
		} else {
			t.pw.Close()
		}
	})
	return t.pr.Read(p)
}

// Write must not be called concurrently with Read. The data must be fully
// written before calling Read.
func (t *Transformer) Write(p []byte) (n int, err error) {
	t.initW.Do(func() {
		if t.pr == nil {
			input, inWriter := io.Pipe()
			output, outWriter := io.Pipe()
			t.pr = output
			t.pw = inWriter
			go handleTransform(input, outWriter, t.mapping)
		} else {
			t.pw = eofReadWriter{}
		}
	})
	return t.pw.Write(p)
}

func handleTransform(in *io.PipeReader, out *io.PipeWriter, mapping Mapping) {
	var lines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()

		// skip empty lines
		if len(line) == 0 {
			continue
		}

		// skip lines that start with a space
		if strings.HasPrefix(line, " ") {
			continue
		}

		// skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Maintain lines in sorted order as they are read.
		idx, _ := slices.BinarySearchFunc(lines, line, compareLines)
		lines = slices.Insert(lines, idx, line)
	}

	if err := scanner.Err(); err != nil {
		out.CloseWithError(err)
		return
	}

	buf := bufio.NewWriter(out)

	defer func() {
		// flush our buffer and close our pipe writer
		out.CloseWithError(buf.Flush())
	}()

	var lastFamily string
	for _, line := range lines {
		family, desc := mapping.get(getFamily(line))
		if family != lastFamily {
			lastFamily = family
			buf.WriteString("# HELP ")
			buf.WriteString(family)
			if desc.Help != "" {
				buf.WriteByte(' ')
				buf.WriteString(desc.Help)
			}
			buf.WriteString("\n# TYPE ")
			buf.WriteString(family)
			buf.WriteByte(' ')
			buf.WriteString(desc.Type.String())
			buf.WriteByte('\n')
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
}

// NewTransformer creates a new Transformer with an optional Mapping.
func NewTransformer(mapping Mapping) *Transformer {
	return &Transformer{
		mapping: mapping,
	}
}

// compareLines compares two metric lines based on their family name
func compareLines(a, b string) int {
	return strings.Compare(getFamily(a), getFamily(b))
}

func getFamily(b string) string {
	// Find either the first { or the first space to extract
	// the family out of the metric line.
	if idx := strings.IndexAny(b, "{ "); idx != -1 {
		return b[:idx]
	}
	return b
}

func normalizeFamily(b string) string {
	// Family names for prom are without the special suffixes
	// for histograms.
	for _, suffix := range [...]string{"_bucket", "_count", "_sum"} {
		if prefix, found := strings.CutSuffix(b, suffix); found {
			return prefix
		}
	}
	return b
}
