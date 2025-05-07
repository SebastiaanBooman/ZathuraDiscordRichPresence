package logging

import (
	"log"
	"os"
	"strings"
)

// Open file for logging (create if it doesn't exist)
func SetupLogger(filePath string) (*os.File, error) {
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	log.SetOutput(logFile)
	return logFile, nil
}

func Error(logMessage string) {
	var builder strings.Builder

	builder.WriteString("[ERROR]: ")
	builder.WriteString(logMessage)

	result := builder.String()
	log.Println(result)
}

func ErrorWithContext(logMessage string, v ...any) {
	var builder strings.Builder

	builder.WriteString("[ERROR]: ")
	builder.WriteString(logMessage)

	result := builder.String()
	log.Printf(result, v)
}

func Warn(logMessage string) {
	var builder strings.Builder

	builder.WriteString("[WARN]: ")
	builder.WriteString(logMessage)

	result := builder.String()
	log.Println(result)
}

func WarnWithContext(logMessage string, v ...any) {
	var builder strings.Builder

	builder.WriteString("[WARN]: ")
	builder.WriteString(logMessage)

	result := builder.String()
	log.Printf(result, v)
}

func LogInfo(logMessage string) {
	var builder strings.Builder

	builder.WriteString("[INFO]: ")
	builder.WriteString(logMessage)

	result := builder.String()
	log.Println(result)
}
