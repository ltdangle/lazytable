package logger

import "os"

type Logger struct {
	file *os.File
}

func NewLogger(path string) *Logger {
	logger := &Logger{}
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	logger.file = file
	return logger
}

func (l *Logger) Info(msg string) {
	msg = msg + "\n"
	_, err := l.file.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}
