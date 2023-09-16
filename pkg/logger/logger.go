package logger

import (
	"fmt"
	"os"
	"strings"
)

var (
	Verbose bool = false
	Logfile string
)

func Success(msg string, args ...interface{}) {
	StdOut("[+] "+msg, args...)
}

func Fail(msg string, args ...interface{}) {
	StdOut("[-] "+msg, args...)
}

func Info(msg string, args ...interface{}) {
	if !Verbose {
		return
	}

	StdOut("[*] "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	StdErr("[!] "+msg, args...)
}

func Critical(msg string, args ...interface{}) {
	StdErr("[!] "+msg, args...)
	os.Exit(1)
}

func StdOut(msg string, args ...interface{}) {
	line := fmt.Sprintf(format(msg), args...)
	logFile(line)
	fmt.Fprint(os.Stdout, line)
}

func StdErr(msg string, args ...interface{}) {
	line := fmt.Sprintf(format(msg), args...)
	logFile(line)
	fmt.Fprint(os.Stderr, line)
}

func logFile(line string) {
	if Logfile != "" {
		file, err := os.OpenFile(Logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			Error("Could not open file '%s': %s", Logfile, err.Error())
			Logfile = "" // Set to null to disable errors in the future
			return
		}
		defer file.Close()

		_, err = file.WriteString(line + "\n")
		if err != nil {
			Error("Could not write to file '%s': %s", Logfile, err.Error())
			Logfile = "" // Set to null to disable errors in the future
			return
		}
	}
}

func format(s string) (out string) {
	out = s

	if !strings.HasSuffix(s, "\n") {
		out = s + "\n"
	}

	return out
}
