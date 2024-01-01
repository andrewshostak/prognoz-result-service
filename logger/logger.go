package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

func SetupLogger(file io.Writer) *zerolog.Logger {
	logger := zerolog.New(zerolog.MultiLevelWriter(file, os.Stderr)).With().Timestamp().Logger()
	return &logger
}

func GetLogFile() (*os.File, error) {
	filename := "app.log"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s file to write logs: %w", filename, err)
	}

	return file, nil
}
