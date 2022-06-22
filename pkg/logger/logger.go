// Copyright 2021 The Karavel Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
)

const (
	debugPrefix = "[DEBUG]"
	infoPrefix  = "[INFO]"
	warnPrefix  = "[WARN]"
	errorPrefix = "[ERROR]"
)

type Logger interface {
	Debug(a ...any)
	Debugf(format string, a ...any)
	Info(a ...any)
	Infof(format string, a ...any)
	Warn(a ...any)
	Warnf(format string, a ...any)
	Error(a ...any)
	Errorf(format string, a ...any)
	Fatal(a ...any)
	Fatalf(format string, a ...any)
	Writer() io.Writer
	Level() Level
	SetLevel(lvl Level)
	SetPalette(palette Palette)
	SetColors(active bool)
}

type logger struct {
	w       io.Writer
	palette palette
	lvl     Level
	colors  bool
	mu      *sync.Mutex
}

func New(lvl Level) Logger {
	return &logger{
		w:       color.Error,
		palette: palettes[PaletteDefault],
		lvl:     lvl,
		mu:      &sync.Mutex{},
	}
}

func (l *logger) Writer() io.Writer {
	return l.w
}

func (l *logger) Level() Level {
	return l.lvl
}

func (l *logger) SetLevel(lvl Level) {
	l.lvl = lvl
}

func (l *logger) SetPalette(palette Palette) {
	l.palette = palettes[palette]
}

func (l *logger) SetColors(active bool) {
	if active {
		l.palette.debug.EnableColor()
		l.palette.info.EnableColor()
		l.palette.warn.EnableColor()
		l.palette.error.EnableColor()
	} else {
		l.palette.debug.DisableColor()
		l.palette.info.DisableColor()
		l.palette.warn.DisableColor()
		l.palette.error.DisableColor()
	}
	l.colors = active
}

func (l *logger) Debug(a ...any) {
	l.output(LvlDebug, a...)
}

func (l *logger) Debugf(format string, a ...any) {
	l.outputf(LvlDebug, format, a...)
}

func (l *logger) Info(a ...any) {
	l.output(LvlInfo, a...)
}

func (l *logger) Infof(format string, a ...any) {
	l.outputf(LvlInfo, format, a...)
}

func (l *logger) Warn(a ...any) {
	l.output(LvlWarn, a...)
}

func (l *logger) Warnf(format string, a ...any) {
	l.outputf(LvlWarn, format, a...)
}

func (l *logger) Error(a ...any) {
	l.output(LvlError, a...)
}

func (l *logger) Errorf(format string, a ...any) {
	l.outputf(LvlError, format, a...)
}

func (l *logger) Fatal(a ...any) {
	l.Error(a...)
	os.Exit(1)
}

func (l *logger) Fatalf(format string, a ...any) {
	l.Errorf(format, a...)
	os.Exit(1)
}

func (l *logger) output(lvl Level, a ...any) {
	if !IsLevelActive(l.lvl, lvl) {
		return
	}

	prefix := false
	if len(a) > 0 {
		prefix = true
	}

	var p func(a ...any) string
	switch lvl {
	case LvlDebug:
		a = append([]any{debugPrefix, " "}, a...)
		p = l.palette.debug.SprintFunc()
	case LvlInfo:
		if prefix && !l.colors {
			a = append([]any{infoPrefix, " "}, a...)
		}
		p = l.palette.info.SprintFunc()
	case LvlWarn:
		a = append([]any{warnPrefix, " "}, a...)
		p = l.palette.warn.SprintFunc()
	case LvlError:
		a = append([]any{errorPrefix, " "}, a...)
		p = l.palette.error.SprintFunc()
	default:
		return
	}
	_, _ = fmt.Fprintln(l.w, p(a...))
}

func (l *logger) outputf(lvl Level, s string, a ...any) {
	if !IsLevelActive(l.lvl, lvl) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	prefix := false
	if s != "" {
		prefix = true
	}

	var p func(format string, a ...any) string
	switch lvl {
	case LvlDebug:
		s = debugPrefix + " " + s
		p = l.palette.debug.SprintfFunc()
	case LvlInfo:
		if prefix && !l.colors {
			s = infoPrefix + " " + s
		}
		p = l.palette.info.SprintfFunc()
	case LvlWarn:
		s = warnPrefix + " " + s
		p = l.palette.warn.SprintfFunc()
	case LvlError:
		s = errorPrefix + " " + s
		p = l.palette.error.SprintfFunc()
	default:
		return
	}

	if len(s) == 0 || s[len(s)-1] != '\n' {
		s += "\n"
	}

	_, _ = fmt.Fprint(l.w, p(s, a...))
}
