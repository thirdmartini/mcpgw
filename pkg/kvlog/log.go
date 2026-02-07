package kvlog

import (
	"fmt"
	"os"

	"github.com/thirdmartini/mcpgw/pkg/txtutils"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[2;37m"
)

const (
	LogError = 1 << iota
	LogWarn
	LogEvent
	LogInfo
	LogHighlight
	LogPrint
	LogDebug

	LogAll = LogError | LogWarn | LogEvent | LogInfo | LogHighlight | LogPrint | LogDebug
)

type Logger interface {
	Printf(s string, v ...interface{}) Logger
	Panicf(s string, v ...interface{}) Logger
	Fatalf(s string, v ...interface{}) Logger
	Infof(s string, v ...interface{}) Logger
	Eventf(s string, v ...interface{}) Logger
	Warnf(s string, v ...interface{}) Logger
	Debugf(s string, v ...interface{}) Logger
	Errorf(s string, v ...interface{}) Logger
	MultiLine(s string) Logger
}

type KVs map[string]interface{}

type Log struct {
	keys      map[string]interface{}
	fmtTime   string
	level     uint32
	component string

	printer func(level uint32, component string, msg string)
}

func (l *Log) Printf(s string, v ...interface{}) Logger {
	return l.emit(LogPrint, s, v...)
}

func (l *Log) Panicf(s string, v ...interface{}) Logger {
	panic(fmt.Sprintf(s, v...))
	return nil
}

func (l *Log) Fatalf(s string, v ...interface{}) Logger {
	l.emit(LogError, s, v...)
	os.Exit(1)
	return nil
}

func (l *Log) Infof(s string, v ...interface{}) Logger {
	return l.emit(LogInfo, s, v...)
}

func (l *Log) Eventf(s string, v ...interface{}) Logger {
	return l.emit(LogEvent, s, v...)
}

func (l *Log) Warnf(s string, v ...interface{}) Logger {
	return l.emit(LogWarn, s, v...)
}

func (l *Log) Debugf(s string, v ...interface{}) Logger {
	return l.emit(LogDebug, s, v...)
}

func (l *Log) Highlightf(s string, v ...interface{}) Logger {
	return l.emit(LogHighlight, s, v...)
}

func (l *Log) Errorf(s string, v ...interface{}) Logger {
	return l.emit(LogError, s, v...)
}

func (l *Log) KVs(kvs KVs) Logger {
	return &Log{
		keys:      kvs,
		fmtTime:   l.fmtTime,
		level:     Default.level,
		printer:   l.printer,
		component: l.component,
	}
}

func (l *Log) WithKVs(kvs KVs) Logger {
	return &Log{
		keys:      kvs,
		fmtTime:   l.fmtTime,
		level:     Default.level,
		printer:   l.printer,
		component: l.component,
	}
}

func (l *Log) MultiLine(s string) Logger {
	lines := txtutils.WrapTextWithNewLines(s, 80)
	for _, line := range lines {
		ColorPrintf(ColorGray, fmt.Sprintf("%28s| %s\n", "", line))
	}
	return l
}

func (l *Log) emit(level uint32, s string, v ...interface{}) Logger {
	if l.level&level == level {
		l.printer(level, l.component, fmt.Sprintf(s, v...))
		for k, v := range l.keys {
			ColorPrintf(ColorGray, fmt.Sprintf("%28s| %s=%v\n", "", k, v))
		}
	}
	return l
}

func WithKV(key string, value interface{}) Logger {
	return WithKVs(map[string]interface{}{key: value})
}

func WithKVs(kvs map[string]interface{}) Logger {
	return &Log{
		keys:      kvs,
		component: "--",
		level:     Default.level,
		fmtTime:   Default.fmtTime,
		printer:   ColorPrinter,
	}
}

func NewLogger(name string) *Log {
	return &Log{
		component: name,
		fmtTime:   Default.fmtTime,
		level:     Default.level,
		printer:   ColorPrinter,
	}
}
