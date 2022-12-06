/*
 * MIT License
 *
 * Copyright (c) 2022 wereliang
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	TimeFormat = "2006-01-02 15:04:05.999"
)

var levelPrefix = [TraceLevel + 1]string{"[E]", "[W]", "[I]", "[D]", "[T]"}

// brush is a color join function
type brush func(string) string

// newBrush return a fix color Brush
func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;31"), // Error              red
	newBrush("1;33"), // Warning            yellow
	newBrush("1;35"), // Informational      Background blue
	newBrush("1;34"), // Debug      		blue
	newBrush("1;32"), // Notice             green
}

func NewSimpleLogger(lv Level, color bool) Logger {
	return &simpleLogger{lv: lv, lg: os.Stdout, color: color}
}

type simpleLogger struct {
	lv    Level
	lg    io.Writer
	color bool
}

func (l *simpleLogger) Trace(format string, args ...interface{}) {
	if l.should(TraceLevel) {
		l.writeMsg(TraceLevel, format, args...)
	}
}

func (l *simpleLogger) Debug(format string, args ...interface{}) {
	if l.should(DebugLevel) {
		l.writeMsg(DebugLevel, format, args...)
	}
}

func (l *simpleLogger) Info(format string, args ...interface{}) {
	if l.should(InfoLevel) {
		l.writeMsg(InfoLevel, format, args...)
	}
}

func (l *simpleLogger) Warn(format string, args ...interface{}) {
	if l.should(WarnLevel) {
		l.writeMsg(WarnLevel, format, args...)
	}
}

func (l *simpleLogger) Error(format string, args ...interface{}) {
	if l.should(ErrorLevel) {
		l.writeMsg(ErrorLevel, format, args...)
	}
}

func (l *simpleLogger) should(lv Level) bool {
	return l.lv >= lv
}

func (l *simpleLogger) writeMsg(lv Level, format string, args ...interface{}) {
	var msg string
	if l.color {
		msg = colors[lv](fmt.Sprintf(format, args...))
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	_, file, line, _ := runtime.Caller(3)
	msg = path.Base(file) + ":" + strconv.Itoa(line) + " " + msg + "\n"
	l.lg.Write(l.formatMsg(time.Now(), lv, msg))
}

func (l *simpleLogger) formatMsg(when time.Time, lv Level, msg string) []byte {
	sb := &strings.Builder{}
	sb.WriteString(when.Format(TimeFormat))
	sb.WriteByte(' ')
	sb.WriteString(levelPrefix[lv])
	sb.WriteByte(' ')
	sb.WriteString(msg)
	return []byte(sb.String())
}
