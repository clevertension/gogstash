package logevent

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tsaikd/KDGoLib/jsonex"
)

type LogEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

const timeFormat = `2006-01-02T15:04:05.999999999Z`

func appendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

func (t *LogEvent) AddTag(tags ...string) {
	for _, tag := range tags {
		ftag := t.Format(tag)
		t.Tags = appendIfMissing(t.Tags, ftag)
	}
}

func (t LogEvent) getJSONMap() map[string]interface{} {
	event := map[string]interface{}{
		"@timestamp": t.Timestamp.UTC().Format(timeFormat),
	}
	if t.Message != "" {
		event["message"] = t.Message
	}
	if len(t.Tags) > 0 {
		event["tags"] = t.Tags
	}
	for key, value := range t.Extra {
		event[key] = value
	}
	return event
}

func (t LogEvent) MarshalJSON() (data []byte, err error) {
	event := t.getJSONMap()
	return jsonex.Marshal(event)
}

func (t LogEvent) MarshalIndent() (data []byte, err error) {
	event := t.getJSONMap()
	return jsonex.MarshalIndent(event, "", "\t")
}

func (t LogEvent) Get(field string) (v interface{}) {
	switch field {
	case "@timestamp":
		v = t.Timestamp
	case "message":
		v = t.Message
	default:
		v = t.Extra[field]
	}
	return
}

func (t LogEvent) GetString(field string) (v string) {
	switch field {
	case "@timestamp":
		v = t.Timestamp.UTC().Format(timeFormat)
	case "message":
		v = t.Message
	default:
		if value, ok := t.Extra[field]; ok {
			v = fmt.Sprintf("%v", value)
		}
	}
	return
}

var (
	retime = regexp.MustCompile(`%{\+([^}]+)}`)
	revar  = regexp.MustCompile(`%{([\w@]+)}`)
)

// FormatWithEnv format string with environment value, ex: %{HOSTNAME}
func FormatWithEnv(text string) (result string) {
	result = text

	matches := revar.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		field := submatches[1]
		value := os.Getenv(field)
		if value != "" {
			result = strings.Replace(result, submatches[0], value, -1)
		}
	}

	return
}

// FormatWithTime format string with current time, ex: %{+2006-01-02}
func FormatWithTime(text string) (result string) {
	result = text

	matches := retime.FindAllStringSubmatch(result, -1)
	for _, submatches := range matches {
		value := time.Now().Format(submatches[1])
		result = strings.Replace(result, submatches[0], value, -1)
	}

	return
}

// Format return string with current time / LogEvent field / ENV, ex: %{hostname}
func (t LogEvent) Format(format string) (out string) {
	out = format

	out = FormatWithTime(out)

	matches := revar.FindAllStringSubmatch(out, -1)
	for _, submatches := range matches {
		field := submatches[1]
		value := t.GetString(field)
		if value != "" {
			out = strings.Replace(out, submatches[0], value, -1)
		}
	}

	out = FormatWithEnv(out)

	return
}
