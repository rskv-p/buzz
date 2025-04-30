package x_str

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rskv-p/buzz/pkg/x_log"
)

// BuildTopicName safely builds and validates a topic name
func BuildTopicName(parts ...string) (string, error) {
	var cleanedParts []string

	for _, part := range parts {
		for _, r := range part {
			if r == '.' || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
				err := fmt.Errorf("invalid character '%c' in part '%s'", r, part)
				x_log.Error("BuildTopicName error:", err)
				return "", err
			}
		}
		cleanedParts = append(cleanedParts, strings.ToLower(part))
	}

	topicName := strings.Join(cleanedParts, ".")
	return topicName, nil
}
