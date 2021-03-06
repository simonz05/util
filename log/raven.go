// Copyright 2014 Simon Zimmermann. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"errors"
	"log"
	"os"

	"github.com/simonz05/util/raven"
)

type ravenLogger struct {
	l   *log.Logger
	sev Level
	dsn string
}

func (r *ravenLogger) Output(calldepth int, s string, sev Level) error {
	if r.sev < sev {
		return nil
	}

	if r.l == nil {
		r.init()
	}

	return r.l.Output(calldepth, s)
}

func (r *ravenLogger) init() {
	c, err := raven.NewClient(r.dsn, "")

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	r.l = log.New(newRavenWriter(c), "", log.Lshortfile)
}

type ravenWriter struct {
	c  *raven.Client
	in chan []byte
}

func newRavenWriter(c *raven.Client) *ravenWriter {
	w := &ravenWriter{
		in: make(chan []byte, 32),
		c:  c,
	}

	go func() {
		w.process()
	}()
	return w
}

func (w *ravenWriter) process() {
	for {
		select {
		case buf, ok := <-w.in:
			if !ok {
				return
			}

			err := w.c.Error(string(buf))

			if err != nil {
				os.Stderr.Write([]byte(err.Error()))
			}
		}
	}
}

// Write implements the io.Writer interface
func (w *ravenWriter) Write(p []byte) (int, error) {
	select {
	case w.in <- p:
		return len(p), nil
	default:
		return 0, errors.New("err chan is full")
	}
}
