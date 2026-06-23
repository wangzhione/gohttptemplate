package handler

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestResponseWriterZipMissingFile(t *testing.T) {
	rec := httptest.NewRecorder()

	err := ResponseWriterZip(context.Background(), rec, "test.zip", []string{"missing-file.zip"})
	if err == nil {
		t.Fatal("expected missing file error")
	}
}
