package config

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type OutputBuffer struct {
	Stdout   strings.Builder
	Stderr   strings.Builder
	OutCount int
	ErrCount int
}

func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{}
}

func (ob *OutputBuffer) Println(a ...any) {
	fmt.Fprintln(&ob.Stdout, a...)
	ob.OutCount += 1
}

func (ob *OutputBuffer) Printf(format string, a ...any) {
	fmt.Fprintf(&ob.Stdout, format, a...)
	ob.OutCount += 1
}

func (ob *OutputBuffer) Errorln(a ...any) {
	fmt.Fprintln(&ob.Stderr, a...)
	ob.ErrCount += 1
}

func (ob *OutputBuffer) Errorf(format string, a ...any) {
	fmt.Fprintf(&ob.Stderr, format, a...)
	ob.ErrCount += 1
}

func (ob *OutputBuffer) Fprintln(dev any, a ...any) {
	switch dev {
	case io.Discard:
		fmt.Fprintln(io.Discard, a...)
	case os.Stdout:
		fmt.Fprintln(&ob.Stdout, a...)
		ob.OutCount += 1
	case os.Stderr:
		fmt.Fprintln(&ob.Stderr, a...)
		ob.ErrCount += 1
	}
}

func (ob *OutputBuffer) Fprintf(dev any, format string, a ...any) {
	switch dev {
	case io.Discard:
		fmt.Fprintf(io.Discard, format, a...)
	case os.Stdout:
		fmt.Fprintf(&ob.Stdout, format, a...)
		ob.OutCount += 1
	case os.Stderr:
		fmt.Fprintf(&ob.Stderr, format, a...)
		ob.ErrCount += 1
	}
}

func (ob *OutputBuffer) Flush() {
	if ob.Stdout.Len() > 0 {
		fmt.Fprint(os.Stdout, ob.Stdout.String())
		ob.OutCount = 0
	}
	if ob.Stderr.Len() > 0 {
		fmt.Fprint(os.Stderr, ob.Stderr.String())
		ob.ErrCount = 0
	}
}
