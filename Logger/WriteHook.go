package Logger

import (
	"github.com/sirupsen/logrus"
	"io"
)

type WriteHook struct {
	writer io.Writer
	levels []logrus.Level
}

// get levels of current hook
func (h * WriteHook) Levels() []logrus.Level {
	return h.levels
}

// Fire is called whenever some logging funcation is called with this hook
// It will format log entry to string and write it to writer
func (h *WriteHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {return err}

	_, err = h.writer.Write([]byte(line))
	return err
}