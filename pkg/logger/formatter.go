package logger

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
)

type ELKFormatter struct{}

func (f *ELKFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logEntry := struct {
		Timestamp string                 `json:"timestamp"`
		Level     string                 `json:"level"`
		Message   string                 `json:"message"`
		Fields    map[string]interface{} `json:"fields,omitempty"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Fields:    entry.Data,
	}

	logData, err := json.Marshal(logEntry)
	if err != nil {
		return nil, err
	}

	logData = append(logData, '\n')
	return logData, nil
}
