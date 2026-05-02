package processor

import (
	"encoding/json"
	"fmt"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

// Parser converts JSON playbook data into model structs.
type Parser struct {
	logger *logger.Log
}

// NewParser builds a parser for supported playbook formats.
func NewParser(_ *config.Config, log *logger.Log) *Parser {
	return &Parser{
		logger: log,
	}
}

// Parse converts raw playbook data using the file extension as format hint.
func (p *Parser) Parse(rawData []byte, filePath string) (*models.Playbook, error) {
	p.logger.Trace("Parsing data, size: %d bytes, format: %s",
		len(rawData), fileFormat(filePath))

	if len(rawData) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var playbook models.Playbook
	var err error

	format := fileFormat(filePath)
	switch format {
	case "json":
		err = p.parseJSON(rawData, &playbook)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", format, err)
	}

	p.logger.Trace("Data parsed: %s, slug: %s",
		playbook.Name, playbook.Slug)

	return &playbook, nil
}

func (p *Parser) parseJSON(data []byte, playbook *models.Playbook) error {
	p.logger.Trace("Parsing JSON")

	if err := json.Unmarshal(data, playbook); err != nil {
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			line := p.findLineNumber(data, syntaxErr.Offset)
			return fmt.Errorf("JSON syntax error at line %d: %w", line, err)
		}
		return fmt.Errorf("JSON unmarshal error: %w", err)
	}

	return nil
}

// findLineNumber reports the 1-based line containing a JSON syntax error offset.
func (p *Parser) findLineNumber(data []byte, offset int64) int {
	if offset >= int64(len(data)) {
		return -1
	}

	line := 1
	for i := int64(0); i < offset; i++ {
		if data[i] == '\n' {
			line++
		}
	}

	return line
}
