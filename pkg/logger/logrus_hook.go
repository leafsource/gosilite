package logger

import (
	"github.com/sirupsen/logrus"
)

type RequestIdHook struct {
	HeaderKey string
}

func (h *RequestIdHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *RequestIdHook) Fire(e *logrus.Entry) error {
	ctx := e.Context
	if ctx == nil {
		return nil
	}
	value := ctx.Value(h.HeaderKey)
	if value != nil {
		e.Data[h.HeaderKey] = value
	}
	return nil
}
