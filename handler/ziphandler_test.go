package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestResponseWriterZipMissingFile(t *testing.T) {
	rec := httptest.NewRecorder()

	err := ResponseWriterZip(context.Background(), rec, "test.zip", []string{"missing-file.zip"})
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestResponseWriterZipDirNestedPaths(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "test.txt"), []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	if err := ResponseWriterZipDir(context.Background(), rec, "test.zip", dir); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}

	expectedNames := []string{"a/", "a/b/", "a/b/test.txt"}
	if len(reader.File) != len(expectedNames) {
		t.Fatalf("expected %d zip entries, got %d", len(expectedNames), len(reader.File))
	}
	for i, file := range reader.File {
		if file.Name != expectedNames[i] {
			t.Fatalf("expected zip entry %q, got %q", expectedNames[i], file.Name)
		}
	}

	file, err := reader.Open("a/b/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Fatalf("unexpected file content %q", content)
	}
}
