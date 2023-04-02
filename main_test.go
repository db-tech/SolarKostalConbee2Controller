package main

import (
	"os"
	"testing"
)

func TestCreateIniFileIfNotExists(t *testing.T) {
	// Create temporary file path for testing
	filepath := "testfile.ini"

	// Clean up temporary file after testing
	defer os.Remove(filepath)

	// Test case 1: File doesn't exist, should create file
	err := CreateIniFileIfNotExists(filepath)
	if err != nil {
		t.Errorf("Test case 1: Expected nil error but got %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("Test case 1: File was not created")
	}

	// Test case 2: File already exists, should not create file
	err = CreateIniFileIfNotExists(filepath)
	if err != nil {
		t.Errorf("Test case 2: Expected nil error but got %v", err)
	}

	// Test case 3: File path is empty, should return error
	err = CreateIniFileIfNotExists("")
	if err == nil {
		t.Error("Test case 3: Expected non-nil error but got nil")
	}

	// Test case 4: File path is invalid, should return error
	err = CreateIniFileIfNotExists("/path/to/nonexistent/directory/testfile.ini")
	if err == nil {
		t.Error("Test case 4: Expected non-nil error but got nil")
	}
}
