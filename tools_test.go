package toolkit

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length of random string returned")
	}
}

func TestTools_UploadedFiles(t *testing.T) {
	// Create a temporary directory for testing
	uploadDir := t.TempDir()

	// Create a new buffer to write our test form data
	var buf bytes.Buffer

	// Create a multipart writer
	writer := multipart.NewWriter(&buf)

	// Create a form file field
	part, err := writer.CreateFormFile("file", "testfile.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	// Write test content to the form file
	_, err = part.Write([]byte("<!DOCTYPE html><html><head><title>Test HTML</title></head><body><h1>Test Content</h1></body></html>"))
	if err != nil {
		t.Fatalf("failed to write to form file: %v", err)
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Create a new request with the form data
	req, err := http.NewRequest("POST", "/upload", &buf)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set the content type to the multipart form content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var testTools Tools
	testTools.MaxFileSize = 1024 * 1024                               // 1 MB
	testTools.AllowedFileTypes = []string{"text/html; charset=utf-8"} // Allow the type that's actually detected

	files, err := testTools.UploadedFiles(req, uploadDir)
	if err != nil {
		t.Fatalf("failed to upload files: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}
