package logger

import (
	"fmt"
	"os"
)

type Logger interface {
	Success(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type ConsoleLogger struct {
	IsVerbose bool
}

func (l *ConsoleLogger) Success(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "\033[32m[OK]\033[0m "+msg+"\n", args...)
}

func (l *ConsoleLogger) Info(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "\033[34m[INFO]\033[0m "+msg+"\n", args...)
}

func (l *ConsoleLogger) Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "\033[31m[ERROR]\033[0m "+msg+"\n", args...)
}

func (l *ConsoleLogger) Debug(msg string, args ...any) {
	if l.IsVerbose {
		fmt.Fprintf(os.Stdout, "\033[90m[DEBUG]\033[0m "+msg+"\n", args...)
	}
}
