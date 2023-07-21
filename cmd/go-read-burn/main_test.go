package main

import (
	"io/ioutil"
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
