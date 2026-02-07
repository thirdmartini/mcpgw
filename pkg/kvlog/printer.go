package kvlog

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var colorMap = map[uint32]string{
	LogError:     ColorRed,
	LogWarn:      ColorYellow,
	LogEvent:     ColorCyan,
	LogInfo:      ColorReset,
	LogHighlight: ColorGreen,
	LogPrint:     ColorGray,
	LogDebug:     ColorGray,
}

var levelMap = map[uint32]string{
	LogError:     "ERROR",
	LogWarn:      "WARN",
	LogEvent:     "EVENT",
	LogInfo:      "INFO",
	LogHighlight: "HILI",
	LogPrint:     "PRINT",
	LogDebug:     "DEBUG",
}

func ColorPrinterClassic(level uint32, component, msg string) {
	color, ok := colorMap[level]
	if !ok {
		color = ColorRed
	}
	now := time.Now() // get this early.
	fmt.Fprintf(os.Stdout, "%s | %s%s%s", now.Format(DefaultTimeFormat), color, msg, ColorReset)
	if !strings.HasSuffix(msg, "\n") {
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func ColorPrinter(level uint32, component, msg string) {
	color, ok := colorMap[level]
	if !ok {
		color = ColorRed
	}
	now := time.Now() // get this early.
	fmt.Fprintf(os.Stdout, "%s | %s%-5.5s%s | %s%s%s", now.Format(DefaultTimeFormat), color, component, ColorReset, color, msg, ColorReset)
	if !strings.HasSuffix(msg, "\n") {
		fmt.Fprintf(os.Stdout, "\n")
	}

}

func BasicPrinter(level uint32, component, msg string) {
	levelString, ok := levelMap[level]
	if !ok {
		levelString = "----"
	}
	now := time.Now() // get this early.
	fmt.Fprintf(os.Stdout, "%s | %-5.5s | %-5.5s | %s", now.Format(DefaultTimeFormat), levelString, component, msg)
	if !strings.HasSuffix(msg, "\n") {
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func ColorPrintf(color string, s string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s%s%s", color, fmt.Sprintf(s, v...), ColorReset)
}
