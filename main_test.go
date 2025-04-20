package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	// Create a temporary file to download to
	tempFile, err := ioutil.TempFile("", "wolfi-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Test successful download
	err = downloadFile(server.URL, tempFile.Name())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify the contents
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Downloaded content doesn't match expected. Got %q, want %q", string(content), "test content")
	}

	// Test failed download (bad URL)
	err = downloadFile("http://localhost:1", tempFile.Name())
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}

	// Test bad status code
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer badServer.Close()

	err = downloadFile(badServer.URL, tempFile.Name())
	if err == nil {
		t.Error("Expected error for bad status code, got nil")
	}

	// Test invalid filepath
	err = downloadFile(server.URL, "/path/that/cant/exist/file.txt")
	if err == nil {
		t.Error("Expected error for invalid filepath, got nil")
	}
}

func TestGetUserCacheDir(t *testing.T) {
	// Save original environment 
	origXDGCacheHome := os.Getenv("XDG_CACHE_HOME")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Could not get user home dir: %v", err)
	}

	// Setup: Reset environment variables
	defer func() {
		if origXDGCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", origXDGCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	testCases := []struct {
		name         string
		mockGOOS     string
		mockXDGCache string
		expected     string
		expectError  bool
	}{
		{
			name:         "Linux with XDG_CACHE_HOME",
			mockGOOS:     "linux",
			mockXDGCache: "/custom/cache/dir",
			expected:     filepath.Join("/custom/cache/dir", cacheSubDir),
		},
		{
			name:         "Linux without XDG_CACHE_HOME",
			mockGOOS:     "linux",
			mockXDGCache: "",
			expected:     filepath.Join(homeDir, ".cache", cacheSubDir),
		},
		{
			name:         "macOS",
			mockGOOS:     "darwin",
			mockXDGCache: "", // Should be ignored on macOS
			expected:     filepath.Join(homeDir, "Library", "Caches", cacheSubDir),
		},
		{
			name:         "Windows",
			mockGOOS:     "windows",
			mockXDGCache: "", // Should be ignored on Windows
			expected:     "", // We cannot easily test the Windows path due to LOCALAPPDATA
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip Windows test in non-Windows environments since we can't easily mock LOCALAPPDATA
			if tc.mockGOOS == "windows" && runtime.GOOS != "windows" {
				t.Skip("Skipping Windows test on non-Windows platform")
			}

			// Mock XDG_CACHE_HOME
			if tc.mockXDGCache != "" {
				os.Setenv("XDG_CACHE_HOME", tc.mockXDGCache)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}

			// Mock GOOS
			if tc.mockGOOS == runtime.GOOS || tc.mockGOOS == "windows" {
				cacheDir, err := getUserCacheDir()
				if tc.expectError {
					if err == nil {
						t.Error("Expected error but got nil")
					}
					return
				}

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if tc.mockGOOS != "windows" && cacheDir != tc.expected {
					t.Errorf("Cache directory mismatch. Got %q, want %q", cacheDir, tc.expected)
				}
			}
		})
	}
}

func TestGetAPKIndexPath(t *testing.T) {
	// Test with provided index path
	providedPath := "testdata/APKINDEX.tar.gz"
	absPath, err := filepath.Abs(providedPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	result, err := getAPKIndexPath(providedPath)
	if err != nil {
		t.Fatalf("getAPKIndexPath failed with provided path: %v", err)
	}
	if result != absPath {
		t.Errorf("Path mismatch. Got %q, want %q", result, absPath)
	}
}