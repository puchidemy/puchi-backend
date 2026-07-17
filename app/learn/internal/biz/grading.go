package biz

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var normalizeSpace = regexp.MustCompile(`\s+`)

// Grade evaluates a learner payload against the exercise answer key.
func Grade(exerciseType string, prompt, answer, payload json.RawMessage) (bool, error) {
	switch exerciseType {
	case "select":
		return gradeSelect(prompt, answer, payload)
	case "match":
		return gradeMatch(answer, payload)
	case "listen", "dictate":
		return gradeTextCompare(answer, payload)
	default:
		return false, fmt.Errorf("unsupported exercise type %q", exerciseType)
	}
}

type selectAnswerKey struct {
	CorrectOptionID *int   `json:"correct_option_id"`
	Correct         string `json:"correct"`
}

type selectPayload struct {
	OptionID json.RawMessage `json:"option_id"`
	Answer   string          `json:"answer"`
}

func gradeSelect(prompt, answer, payload json.RawMessage) (bool, error) {
	var key selectAnswerKey
	if err := json.Unmarshal(answer, &key); err != nil {
		return false, fmt.Errorf("parse select answer: %w", err)
	}
	var p selectPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return false, fmt.Errorf("parse select payload: %w", err)
	}

	if key.CorrectOptionID != nil {
		got, err := parseOptionID(p.OptionID)
		if err != nil {
			return false, err
		}
		return got == *key.CorrectOptionID, nil
	}

	if key.Correct != "" {
		if p.Answer != "" {
			return normalizeText(key.Correct) == normalizeText(p.Answer), nil
		}
		got, err := optionTextAtIndex(prompt, p.OptionID)
		if err != nil {
			return false, err
		}
		return normalizeText(key.Correct) == normalizeText(got), nil
	}

	return false, fmt.Errorf("select answer key missing correct_option_id or correct")
}

func parseOptionID(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("option_id required")
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, fmt.Errorf("parse option_id: %w", err)
	}
	return strconv.Atoi(s)
}

func optionTextAtIndex(prompt, optionIDRaw json.RawMessage) (string, error) {
	idx, err := parseOptionID(optionIDRaw)
	if err != nil {
		return "", err
	}
	var doc struct {
		Options []string `json:"options"`
	}
	if err := json.Unmarshal(prompt, &doc); err != nil {
		return "", fmt.Errorf("parse select prompt: %w", err)
	}
	if idx < 0 || idx >= len(doc.Options) {
		return "", fmt.Errorf("option_id out of range")
	}
	return doc.Options[idx], nil
}

type matchBody struct {
	Pairs [][2]string `json:"pairs"`
}

func gradeMatch(answer, payload json.RawMessage) (bool, error) {
	var expected, got matchBody
	if err := json.Unmarshal(answer, &expected); err != nil {
		return false, fmt.Errorf("parse match answer: %w", err)
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		return false, fmt.Errorf("parse match payload: %w", err)
	}
	return pairSetsEqual(expected.Pairs, got.Pairs), nil
}

func pairSetsEqual(a, b [][2]string) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
outer:
	for _, p := range a {
		for i, q := range b {
			if used[i] {
				continue
			}
			if p[0] == q[0] && p[1] == q[1] {
				used[i] = true
				continue outer
			}
		}
		return false
	}
	return true
}

type textBody struct {
	Text string `json:"text"`
}

func gradeTextCompare(answer, payload json.RawMessage) (bool, error) {
	var expected, got textBody
	if err := json.Unmarshal(answer, &expected); err != nil {
		return false, fmt.Errorf("parse text answer: %w", err)
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		return false, fmt.Errorf("parse text payload: %w", err)
	}
	return normalizeText(expected.Text) == normalizeText(got.Text), nil
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = normalizeSpace.ReplaceAllString(s, " ")
	re := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	s = re.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
