package promhttp

import (
	"bufio"
	"io"
	"slices"
	"strings"
)

type Type uint8

const (
	Untyped Type = iota
	Counter
	Gauge
	Histogram
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
	case Info:
		return "info"
	default:
		panic("unknown Type")
	}
}

type Desc struct {
	Type Type
	Help string
}

type Mapping map[string]Desc

type Transformer struct {
	pr      *io.PipeReader
	pw      *io.PipeWriter
	scanner *bufio.Scanner
	mapping Mapping
}

func (t *Transformer) Read(p []byte) (n int, err error) {
	return t.pr.Read(p)
}

func (t *Transformer) run() {
	defer t.pw.Close()

	var lines []string
	for t.scanner.Scan() {
		line := t.scanner.Text()

		// skip comment lines
		if strings.HasPrefix(line, "# ") {
			continue
		}

		// skip empty lines
		if strings.HasPrefix(line, " ") {
			continue
		}

		// Maintain lines in sorted order as they are read.
		idx, _ := slices.BinarySearchFunc(lines, line, compareLines)
		lines = slices.Insert(lines, idx, line)
	}

	out := t.pw
	var lastFamily string
	for _, line := range lines {
		family := getFamily(line)
		if family != lastFamily {
			lastFamily = family
			m := t.mapping[family]
			io.WriteString(out, "# HELP ")
			io.WriteString(out, family)
			if m.Help != "" {
				io.WriteString(out, " "+m.Help)
			}
			io.WriteString(out, "\n# TYPE ")
			io.WriteString(out, m.Type.String()+"\n")
		}
		io.WriteString(out, line+"\n")
	}
}

func NewTransformer(in io.Reader, mapping Mapping) *Transformer {
	pr, pw := io.Pipe()
	t := &Transformer{
		pr:      pr,
		pw:      pw,
		scanner: bufio.NewScanner(in),
		mapping: mapping,
	}
	go t.run()
	return t
}

// compareLines compares two metric lines based on their family name
func compareLines(a, b string) int {
	return strings.Compare(getFamily(a), getFamily(b))
}

func getFamily(b string) string {
	// Find either the first { or the first space to extract
	// the family out of the metric line.
	if idx := strings.IndexByte(b, '{'); idx != -1 {
		b = b[:idx]
	} else if idx := strings.IndexByte(b, ' '); idx != -1 {
		b = b[:idx]
	}

	// Family names for prom are without the special suffixes
	// for histograms.
	switch {
	case strings.HasSuffix(b, "_bucket"):
		b = strings.TrimSuffix(b, "_bucket")
	case strings.HasSuffix(b, "_count"):
		b = strings.TrimSuffix(b, "_count")
	case strings.HasSuffix(b, "_sum"):
		b = strings.TrimSuffix(b, "_sum")
	}

	return b
}
