package config

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewLogger(viper *viper.Viper) *logrus.Logger {
	log := logrus.New()

	log.SetLevel(logrus.Level(viper.GetInt32("log.level")))
	// log.SetFormatter(&logrus.JSONFormatter{}) // Example: Custom formatter to control field order
	log.SetFormatter(&customLogrusFormatter{})

	return log
}

type customLogrusFormatter struct{}

func (f *customLogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	type logEntry struct {
		Time  string                 `json:"time"`
		Level string                 `json:"level"`
		Msg   string                 `json:"message"`
		Data  map[string]interface{} `json:"data,omitempty"`
	}
	entryStruct := logEntry{
		Time:  entry.Time.Format("2006-01-02T15:04:05Z07:00"),
		Level: entry.Level.String(),
		Msg:   entry.Message,
	}
	if len(entry.Data) > 0 {
		entryStruct.Data = entry.Data
	}
	b, err := json.Marshal(entryStruct)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
