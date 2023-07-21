package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateDBDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case: Directory does not exist
	dbPath := tempDir + "/db/secrets.db"
	err = createDBDir(dbPath)
	if err != nil {
		t.Errorf("Failed to create directory: %v", err)
	}

	// Check if directory was created
	_, err = os.Stat(tempDir + "/db")
	if os.IsNotExist(err) {
		t.Errorf("Directory was not created")
	}

	// Test case: Directory already exists
	err = createDBDir(dbPath)
	if err != nil {
		t.Errorf("Failed when directory already exists: %v", err)
	}
}

func TestIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(IndexHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `Home`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
