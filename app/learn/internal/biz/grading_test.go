package biz

import (
	"encoding/json"
	"testing"
)

func TestGradeSelect_Correct(t *testing.T) {
	prompt := json.RawMessage(`{"options":[{"id":0,"text":"Hello"},{"id":1,"text":"Bye"}]}`)
	answer := json.RawMessage(`{"correct_option_id":0}`)
	payload := json.RawMessage(`{"option_id":0}`)

	ok, err := Grade("select", prompt, answer, payload)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if !ok {
		t.Fatal("expected correct grade")
	}
}

func TestGradeSelect_Wrong(t *testing.T) {
	prompt := json.RawMessage(`{"options":[{"id":0,"text":"Hello"},{"id":1,"text":"Bye"}]}`)
	answer := json.RawMessage(`{"correct_option_id":0}`)
	payload := json.RawMessage(`{"option_id":1}`)

	ok, err := Grade("select", prompt, answer, payload)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if ok {
		t.Fatal("expected incorrect grade")
	}
}

func TestGradeMatch_OrderIndependent(t *testing.T) {
	answer := json.RawMessage(`{"pairs":[["I","T\u00f4i"],["You","B\u1ea1n"]]}`)
	payload := json.RawMessage(`{"pairs":[["You","B\u1ea1n"],["I","T\u00f4i"]]}`)

	ok, err := Grade("match", nil, answer, payload)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if !ok {
		t.Fatal("expected correct grade for order-independent match")
	}
}

func TestGradeListen_NormalizedCompare(t *testing.T) {
	answer := json.RawMessage(`{"text":"Hello!"}`)
	payload := json.RawMessage(`{"text":"  hello "}`)

	ok, err := Grade("listen", nil, answer, payload)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if !ok {
		t.Fatal("expected normalized text match")
	}
}
