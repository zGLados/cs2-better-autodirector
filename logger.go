package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	// VerboseMode controls detailed logging
	VerboseMode = false
	// LogFile for writing detailed logs
	LogFile *os.File
	// fileLogger writes only to file
	fileLogger *log.Logger
	// mainLogger writes to both console and file
	mainLogger *log.Logger
)

// InitLogging sets up logging based on verbose mode
func InitLogging(verbose bool) error {
	VerboseMode = verbose

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create log file with timestamp (always)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := filepath.Join("logs", fmt.Sprintf("autoobserver_%s.log", timestamp))

	var err error
	LogFile, err = os.Create(logFileName)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}

	// Create loggers
	fileLogger = log.New(LogFile, "", log.Ltime)
	mainLogger = log.New(io.MultiWriter(os.Stdout, LogFile), "", log.Ltime)

	if verbose {
		fmt.Println("Verbose mode enabled - showing all logs in console AND file")
		fmt.Printf("Log file: %s\n", logFileName)
	} else {
		fmt.Printf("Detailed logs will be written to: %s\n", logFileName)
	}

	return nil
}

// CloseLogging closes the log file if open
func CloseLogging() {
	if LogFile != nil {
		LogFile.Close()
	}
}

// LogInfo logs a message that's always shown (console + file)
func LogInfo(format string, args ...interface{}) {
	mainLogger.Printf("[INFO] "+format, args...)
}

// LogVerbose logs a message (console+file in verbose mode, only file in normal mode)
func LogVerbose(format string, args ...interface{}) {
	if VerboseMode {
		mainLogger.Printf(format, args...)
	} else {
		fileLogger.Printf(format, args...)
	}
}

// LogDebug logs debug information (console+file in verbose mode, only file in normal mode)
func LogDebug(format string, args ...interface{}) {
	if VerboseMode {
		mainLogger.Printf("[DEBUG] "+format, args...)
	} else {
		fileLogger.Printf("[DEBUG] "+format, args...)
	}
}
