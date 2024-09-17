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
	return l.getMessagesForLevels(level)
}

// GetErrorMessages returns all log messages with error, fatal and panic level
// If you only want a single of those levels, use GetMessagesForLevel instead
func (l *TestLogger) GetErrorMessages() ([]LogMessage, error) {
	return l.getMessagesForLevels(logger.Level(zerolog.ErrorLevel), logger.Level(zerolog.FatalLevel), logger.Level(zerolog.PanicLevel))
}

func (l *TestLogger) getMessagesForLevels(levels ...logger.Level) ([]LogMessage, error) {
	messages, err := l.GetLogMessages()
	if err != nil {
		return nil, err
	}

	result := []LogMessage{}

	for _, log := range messages {
		shouldContinue := true
		for _, lvl := range levels {
			if logger.Level(log.Level) == lvl {
				shouldContinue = false
			}
		}
		if shouldContinue {
			continue
		}

		result = append(result, log)
	}
	return result, nil
}
