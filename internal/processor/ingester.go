package processor

import (
	"fmt"
	"os"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
)

// Ingester reads playbook files from disk.
type Ingester struct {
	logger *logger.Log
}

// NewIngester builds an ingester for supported playbook formats.
func NewIngester(_ *config.Config, log *logger.Log) *Ingester {
	return &Ingester{
		logger: log,
	}
}

const maxPlaybookSize = 10 * 1024 * 1024 // 10 MB

// IngestFile returns the raw contents of a supported playbook file.
func (i *Ingester) IngestFile(filePath string) ([]byte, error) {
	i.logger.Trace("Reading file: %s", filePath)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > maxPlaybookSize {
		return nil, fmt.Errorf("playbook file too large: %d bytes (max %d)", info.Size(), maxPlaybookSize)
	}

	if !i.isSupportedFormat(filePath) {
		return nil, fmt.Errorf("unsupported file format: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	i.logger.Trace("File read, size: %d bytes, format: %s",
		len(data), fileFormat(filePath))

	return data, nil
}

var supportedFormats = map[string]bool{
	"json": true,
}

func (i *Ingester) isSupportedFormat(filePath string) bool {
	ext := fileFormat(filePath)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	return supportedFormats[ext]
}
