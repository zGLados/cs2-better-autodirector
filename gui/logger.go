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
	// LogLevel controls logging depth (0=Info, 1=Verbose, 2=Debug/Trace)
	LogLevel = 0
	// LogFile for writing detailed logs
	LogFile *os.File
	// fileLogger writes only to file
	fileLogger *log.Logger
	// mainLogger writes to both console and file
	mainLogger *log.Logger
)

// InitLogging sets up logging based on verbose mode
func InitLogging(level int) error {
	LogLevel = level

	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(getConfigDir(), "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create log file with timestamp (always)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := filepath.Join(logsDir, fmt.Sprintf("autodirector_%s.log", timestamp))

	var err error
	LogFile, err = os.Create(logFileName)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}

	// Create loggers
	fileLogger = log.New(LogFile, "", log.Ltime)
	mainLogger = log.New(io.MultiWriter(os.Stdout, LogFile), "", log.Ltime)

	if LogLevel > 0 {
		fmt.Println("=====================================")
		fmt.Printf("📊 Logging Level: %d\n", LogLevel)
		fmt.Println("  → Console: Shows detailed logs")
		fmt.Println("  → File:    Shows ALL logs")
		fmt.Printf("  → Log file: %s\n", logFileName)
		fmt.Println("=====================================")
	} else {
		fmt.Println("=====================================")
		fmt.Println("📋 Info Mode")
		fmt.Println("  → Console: Shows IMPORTANT events only")
		fmt.Println("  → File:    Shows ALL detailed logs")
		fmt.Printf("  → Log file: %s\n", logFileName)
		fmt.Println("  → Tip: Run with -v or -vv for more details")
		fmt.Println("=====================================")
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
	if LogLevel >= 1 {
		mainLogger.Printf(format, args...)
	} else {
		fileLogger.Printf(format, args...)
	}
}

// LogDebug logs debug information (console+file in verbose mode, only file in normal mode)
func LogDebug(format string, args ...interface{}) {
	if LogLevel >= 2 {
		mainLogger.Printf("[DEBUG] "+format, args...)
	} else {
		fileLogger.Printf("[DEBUG] "+format, args...)
	}
}
