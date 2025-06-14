package pkg

import (
	"encoding/json"
	"goilerplate/config"
	"runtime"

	"github.com/sirupsen/logrus"
)

func NewLogger(cfg *config.Config) *logrus.Logger {
	log := logrus.New()

	log.SetLevel(logrus.Level(cfg.Log.Level))
	log.SetFormatter(&customLogrusFormatter{})
	log.AddHook(&callerHook{})
	log.SetReportCaller(false)

	return log
}

type callerHook struct{}

func (h *callerHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (h *callerHook) Fire(entry *logrus.Entry) error {
	if pc, file, line, ok := runtime.Caller(8); ok {
		funcName := runtime.FuncForPC(pc).Name()
		entry.Caller = &runtime.Frame{
			File:     file,
			Line:     line,
			Function: funcName,
		}
	}
	return nil
}

type customLogrusFormatter struct{}

func (f *customLogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	type logEntry struct {
		Time  string         `json:"time"`
		Level string         `json:"level"`
		Msg   string         `json:"message"`
		File  string         `json:"file,omitempty"`
		Line  int            `json:"line,omitempty"`
		Func  string         `json:"func,omitempty"`
		Data  map[string]any `json:"data,omitempty"`
	}
	entryStruct := logEntry{
		Time:  entry.Time.Format("2006-01-02T15:04:05Z07:00"),
		Level: entry.Level.String(),
		Msg:   entry.Message,
	}
	if entry.Caller != nil {
		entryStruct.File = entry.Caller.File
		entryStruct.Line = entry.Caller.Line
		entryStruct.Func = entry.Caller.Function
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
