package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestResponseWriterMessageEscapesJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	message := "bad \"quote\"\nnext"

	ResponseWriterMessage(rec, message)

	var resp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if resp.Message != message {
		t.Fatalf("expected message %q, got %q", message, resp.Message)
	}
}
