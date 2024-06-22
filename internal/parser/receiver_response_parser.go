package parser

import (
	"errors"
	"github.com/mguley/web-scraper-v1/internal/model"
	"strings"
)

// ReceiverResponseParser is an implementation of the Parser interface for ReceiverResponse.
type ReceiverResponseParser struct{}

// NewReceiverResponseParser creates a new instance of ReceiverResponseParser.
func NewReceiverResponseParser() *ReceiverResponseParser {
	return &ReceiverResponseParser{}
}

// Parse converts a string containing response data into a ReceiverResponse struct.
func (parser *ReceiverResponseParser) Parse(htmlContent string) (*model.ReceiverResponse, error) {
	lines := strings.Split(htmlContent, "\n")
	if len(lines) < 3 {
		return nil, errors.New("invalid response format")
	}

	response := &model.ReceiverResponse{}

	for _, line := range lines {
		if strings.HasPrefix(line, "Received User-Agent: ") {
			response.UserAgent = strings.TrimPrefix(line, "Received User-Agent: ")
		} else if strings.HasPrefix(line, "IP Address: ") {
			response.IPAddress = strings.TrimPrefix(line, "IP Address: ")
		} else if strings.HasPrefix(line, "Forwarded Host: ") {
			response.ForwardedHost = strings.TrimPrefix(line, "Forwarded Host: ")
		}
	}

	return response, nil
}
