package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"chainguard.dev/apko/pkg/apk/apk"
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
	// Test with provided local file path
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

func TestIsURL(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz", true},
		{"/path/to/file.tar.gz", false},
		{"file:///path/to/file.tar.gz", false},
		{"APKINDEX.tar.gz", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isURL(tc.input)
			if result != tc.expected {
				t.Errorf("isURL(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetAPKIndexPathFromURL(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test APKINDEX content"))
	}))
	defer server.Close()

	// Test URL handling
	testURL := server.URL + "/APKINDEX.tar.gz"

	// Call getAPKIndexPath with the URL
	result, err := getAPKIndexPath(testURL)
	if err != nil {
		t.Fatalf("getAPKIndexPath failed with URL: %v", err)
	}

	// Verify the result is a local file path
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist at %s", result)
	} else if err != nil {
		t.Errorf("Error checking downloaded file: %v", err)
	}

	// Verify the URL was processed correctly
	if !strings.Contains(result, "APKINDEX_") {
		t.Errorf("Expected result to contain 'APKINDEX_' (hash of URL), got %s", result)
	}

	// Verify the content was downloaded correctly
	content, err := ioutil.ReadFile(result)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != "test APKINDEX content" {
		t.Errorf("Downloaded content doesn't match expected. Got %q, want %q",
			string(content), "test APKINDEX content")
	}
}

func TestMultiStringFlag(t *testing.T) {
	// Test empty flag
	var flag multiStringFlag
	if flag.String() != "" {
		t.Errorf("Expected empty string, got %q", flag.String())
	}

	// Test adding one value
	err := flag.Set("value1")
	if err != nil {
		t.Fatalf("Unexpected error setting flag: %v", err)
	}
	if flag.String() != "value1" {
		t.Errorf("Expected 'value1', got %q", flag.String())
	}

	// Test adding multiple values
	err = flag.Set("value2")
	if err != nil {
		t.Fatalf("Unexpected error setting flag: %v", err)
	}
	err = flag.Set("value3")
	if err != nil {
		t.Fatalf("Unexpected error setting flag: %v", err)
	}

	// Check the String() method
	expected := "value1, value2, value3"
	if flag.String() != expected {
		t.Errorf("Expected %q, got %q", expected, flag.String())
	}

	// Check flag length
	if len(flag) != 3 {
		t.Errorf("Expected 3 values, got %d", len(flag))
	}

	// Check flag values directly
	expectedValues := []string{"value1", "value2", "value3"}
	for i, v := range flag {
		if v != expectedValues[i] {
			t.Errorf("Value at index %d mismatch. Expected %q, got %q", i, expectedValues[i], v)
		}
	}
}

func TestIntegrationMultipleIndexes(t *testing.T) {
	// This test simulates loading multiple APKINDEX files and checks that the merging works correctly
	// It requires the APKINDEX files to exist in the testdata directory
	// If they don't exist, this test will be skipped

	// Would create temporary directory for test data
	// testDir, err := ioutil.TempDir("", "wolfi-mcp-test")
	// if err != nil {
	//     t.Fatalf("Failed to create temp directory: %v", err)
	// }
	// defer os.RemoveAll(testDir)

	// Would create two simulated APKINDEX.tar.gz files with different package versions
	// index1Path := filepath.Join(testDir, "index1.tar.gz")
	// index2Path := filepath.Join(testDir, "index2.tar.gz")

	// We'll skip file creation and actual loading since that would require creating complex APKINDEX files
	// This is just a placeholder for a full integration test
	t.Skip("Integration test for multiple APKINDEX loading requires actual APKINDEX files")

	// In a real test, we would:
	// 1. Create valid APKINDEX files with known packages/versions
	// 2. Run the loading and merging process
	// 3. Verify the resulting repository has the correct packages after merging
}

func TestMergePackages(t *testing.T) {
	testCases := []struct {
		name     string
		existing []*apk.Package
		new      []*apk.Package
		expected []*apk.Package
	}{
		{
			name:     "Empty existing packages",
			existing: []*apk.Package{},
			new: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
			},
		},
		{
			name: "Empty new packages",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
			},
			new: []*apk.Package{},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
			},
		},
		{
			name: "New package not in existing",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
			},
			new: []*apk.Package{
				{Name: "pkg2", Version: "2.0.0"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
			},
		},
		{
			name: "Higher version in new",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
			},
			new: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
			},
		},
		{
			name: "Lower version in new",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
			},
			new: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
			},
		},
		{
			name: "Same version, take new",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0", Description: "Old description"},
			},
			new: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0", Description: "New description"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0", Description: "New description"},
			},
		},
		{
			name: "Complex mix of scenarios",
			existing: []*apk.Package{
				{Name: "pkg1", Version: "1.0.0"},
				{Name: "pkg2", Version: "2.0.0"},
				{Name: "pkg3", Version: "3.1.0"},
				{Name: "pkg4", Version: "4.0.0", Description: "Old pkg4"},
			},
			new: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
				{Name: "pkg2", Version: "1.9.0"},
				{Name: "pkg4", Version: "4.0.0", Description: "New pkg4"},
				{Name: "pkg5", Version: "5.0.0"},
			},
			expected: []*apk.Package{
				{Name: "pkg1", Version: "1.1.0"},
				{Name: "pkg2", Version: "2.0.0"},
				{Name: "pkg3", Version: "3.1.0"},
				{Name: "pkg4", Version: "4.0.0", Description: "New pkg4"},
				{Name: "pkg5", Version: "5.0.0"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mergePackages(tc.existing, tc.new)

			// Since the order of packages in the result is not guaranteed
			// (it depends on map iteration order), we need to check for equality
			// by first converting both to maps for comparison

			expectedMap := make(map[string]*apk.Package)
			for _, pkg := range tc.expected {
				expectedMap[pkg.Name] = pkg
			}

			resultMap := make(map[string]*apk.Package)
			for _, pkg := range result {
				resultMap[pkg.Name] = pkg
			}

			// Check number of packages
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d packages, got %d", len(tc.expected), len(result))
			}

			// Check each package
			for name, expectedPkg := range expectedMap {
				resultPkg, exists := resultMap[name]
				if !exists {
					t.Errorf("Expected package %s not found in result", name)
					continue
				}

				if resultPkg.Version != expectedPkg.Version {
					t.Errorf("Package %s version mismatch. Expected %s, got %s",
						name, expectedPkg.Version, resultPkg.Version)
				}

				if expectedPkg.Description != "" && resultPkg.Description != expectedPkg.Description {
					t.Errorf("Package %s description mismatch. Expected %s, got %s",
						name, expectedPkg.Description, resultPkg.Description)
				}
			}
		})
	}
}
