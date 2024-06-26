package testlogger

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/rs/zerolog"

	"github.com/openmfp/golang-commons/logger"
)

type TestLogger struct {
	*logger.Logger
	buffer *bytes.Buffer
}

// New returns a logger with an in memory buffer containing log messages for use in tests
func New() *TestLogger {
	buf := &bytes.Buffer{}
	cfg := logger.DefaultConfig()
	cfg.Level = "debug"
	cfg.Output = buf
	l, _ := logger.New(cfg)

	return &TestLogger{
		Logger: l,
		buffer: buf,
	}
}

type LogMessage struct {
	Message    string                 `json:"message"`
	Level      zerolog.Level          `json:"level"`
	Service    string                 `json:"service"`
	Error      *string                `json:"error"`
	Attributes map[string]interface{} `json:"-"`
}

func (l *TestLogger) GetLogMessages() ([]LogMessage, error) {
	result := make([]LogMessage, 0)
	logString := l.buffer.String()
	messages := strings.Split(logString, "\n")
	for _, message := range messages {
		if message == "" {
			continue
		}
		logMessage := LogMessage{}
		err := json.Unmarshal([]byte(message), &logMessage)
		if err != nil {
			return nil, err
		}

		attributes := map[string]interface{}{}
		err = json.Unmarshal([]byte(message), &attributes)
		if err != nil {
			return nil, err
		}
		logMessage.Attributes = attributes

		result = append(result, logMessage)
	}

	return result, nil
}

func (l *TestLogger) GetMessagesForLevel(level logger.Level) ([]LogMessage, error) {
	messages, err := l.GetLogMessages()
	if err != nil {
		return nil, err
	}

	result := []LogMessage{}

	for _, log := range messages {
		if logger.Level(log.Level) != level {
			continue
		}
		result = append(result, log)
	}
	return result, nil
}

func (l *TestLogger) GetErrorMessages() ([]LogMessage, error) {
	return l.GetMessagesForLevel(logger.Level(zerolog.ErrorLevel))
}
