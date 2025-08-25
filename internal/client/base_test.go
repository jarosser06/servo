package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/servo/servo/pkg"
)

// TestBaseClientInfo tests the BaseClientInfo struct directly
func TestBaseClientInfo(t *testing.T) {
	info := BaseClientInfo{
		Name:        "test",
		Description: "Test Client",
		Platforms:   []string{"darwin", "linux"},
	}

	if info.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", info.Name)
	}

	if info.Description != "Test Client" {
		t.Errorf("expected description 'Test Client', got '%s'", info.Description)
	}

	if len(info.Platforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(info.Platforms))
	}
}

func TestIsPlatformSupportedForPlatforms(t *testing.T) {
	// Test current platform support based on runtime.GOOS
	supported := IsPlatformSupportedForPlatforms([]string{"darwin", "linux"})

	// Should return true if current platform is in the list, false otherwise
	expectedSupported := false
	for _, platform := range []string{"darwin", "linux"} {
		if platform == runtime.GOOS {
			expectedSupported = true
			break
		}
	}

	if supported != expectedSupported {
		t.Errorf("expected platform support %v for %s, got %v", expectedSupported, runtime.GOOS, supported)
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing file
	existingFile := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	if !FileExists(existingFile) {
		t.Errorf("FileExists should return true for existing file")
	}

	// Test non-existing file
	nonExistingFile := filepath.Join(tmpDir, "nonexistent.txt")
	if FileExists(nonExistingFile) {
		t.Errorf("FileExists should return false for non-existing file")
	}
}

func TestExecutableExists(t *testing.T) {
	// Test with a command that should exist on most systems
	if !ExecutableExists("echo") {
		t.Logf("echo command not found, this might be expected in some environments")
	}

	// Test with a command that definitely shouldn't exist
	if ExecutableExists("definitely-nonexistent-command-12345") {
		t.Errorf("ExecutableExists should return false for non-existent command")
	}
}

func TestGetExecutableVersion(t *testing.T) {
	// Test with echo command (should be available on most systems)
	version, err := GetExecutableVersion("echo", "--version")
	if err != nil {
		t.Logf("echo --version failed, might not be supported: %v", err)
	} else if version == "" {
		t.Logf("echo --version returned empty string")
	}

	// Test with non-existent command
	_, err = GetExecutableVersion("definitely-nonexistent-command-12345", "--version")
	if err == nil {
		t.Errorf("GetExecutableVersion should return error for non-existent command")
	}
}

func TestExpandPath(t *testing.T) {
	// Test absolute path (should remain unchanged)
	absPath := "/absolute/path"
	expanded, err := ExpandPath(absPath)
	if err != nil {
		t.Errorf("ExpandPath should not error for absolute path: %v", err)
	}
	if expanded != absPath {
		t.Errorf("expected '%s', got '%s'", absPath, expanded)
	}

	// Test relative path
	relPath := "relative/path"
	expanded, err = ExpandPath(relPath)
	if err != nil {
		t.Errorf("ExpandPath should not error for relative path: %v", err)
	}
	if !filepath.IsAbs(expanded) {
		t.Errorf("ExpandPath should return absolute path, got '%s'", expanded)
	}

	// Test tilde path (if HOME is set)
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		tildeLink := "~/test/path"
		expanded, err = ExpandPath(tildeLink)
		if err != nil {
			t.Errorf("ExpandPath should not error for tilde path: %v", err)
		}
		expected := filepath.Join(homeDir, "test/path")
		if expanded != expected {
			t.Errorf("expected '%s', got '%s'", expected, expanded)
		}
	}
}

func TestEnsureDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "new", "nested", "directory")

	// Test creating nested directories
	err := EnsureDirectory(testDir)
	if err != nil {
		t.Errorf("EnsureDirectory should not error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("EnsureDirectory should create directory")
	}

	// Test with existing directory (should not error)
	err = EnsureDirectory(testDir)
	if err != nil {
		t.Errorf("EnsureDirectory should not error for existing directory: %v", err)
	}
}

func TestReadJSONFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test valid JSON file
	testFile := filepath.Join(tmpDir, "test.json")
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	data, _ := json.Marshal(testData)
	os.WriteFile(testFile, data, 0644)

	var result map[string]interface{}
	err := ReadJSONFile(testFile, &result)
	if err != nil {
		t.Errorf("ReadJSONFile should not error for valid JSON: %v", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("expected key1='value1', got '%v'", result["key1"])
	}

	// Test non-existent file
	err = ReadJSONFile(filepath.Join(tmpDir, "nonexistent.json"), &result)
	if err == nil {
		t.Errorf("ReadJSONFile should error for non-existent file")
	}

	// Test invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	err = ReadJSONFile(invalidFile, &result)
	if err == nil {
		t.Errorf("ReadJSONFile should error for invalid JSON")
	}
}

func TestWriteJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.json")

	testData := map[string]interface{}{
		"test":   "data",
		"number": 123,
	}

	// Test writing JSON
	err := WriteJSONFile(testFile, testData)
	if err != nil {
		t.Errorf("WriteJSONFile should not error: %v", err)
	}

	// Verify file was written correctly
	var result map[string]interface{}
	err = ReadJSONFile(testFile, &result)
	if err != nil {
		t.Errorf("ReadJSONFile should work on written file: %v", err)
	}

	if result["test"] != "data" {
		t.Errorf("expected test='data', got '%v'", result["test"])
	}
}

func TestServerConfigsToMCP(t *testing.T) {
	serverConfigs := []pkg.ServerConfig{
		{
			Version: "1.0.0",
			Clients: []string{"vscode", "cursor"},
		},
		{
			Version: "2.0.0",
			Clients: []string{"claude-desktop"},
		},
	}

	mcpConfigs := ServerConfigsToMCP(serverConfigs)

	if len(mcpConfigs) != 2 {
		t.Errorf("expected 2 MCP configs, got %d", len(mcpConfigs))
	}

	// Check that configs were created
	if len(mcpConfigs) == 0 {
		t.Errorf("expected at least one MCP config")
	}

	// Basic validation - the exact structure depends on implementation
	for name, config := range mcpConfigs {
		if name == "" {
			t.Errorf("config name should not be empty")
		}
		if config.Command == "" {
			t.Logf("config command is empty for %s", name)
		}
	}
}

func TestMergeServerConfigs(t *testing.T) {
	existing := map[string]pkg.MCPServerConfig{
		"server1": {
			Command: "existing-command",
			Args:    []string{"--existing"},
		},
	}

	new := map[string]pkg.MCPServerConfig{
		"server2": {
			Command: "new-command",
			Args:    []string{"--new"},
		},
		"server1": { // This should overwrite
			Command: "updated-command",
			Args:    []string{"--updated"},
		},
	}

	result := MergeServerConfigs(existing, new)

	if len(result) != 2 {
		t.Errorf("expected 2 servers, got %d", len(result))
	}

	// Check that server1 was overwritten
	if result["server1"].Command != "updated-command" {
		t.Errorf("expected server1 command 'updated-command', got '%s'", result["server1"].Command)
	}

	// Check that server2 was added
	if result["server2"].Command != "new-command" {
		t.Errorf("expected server2 command 'new-command', got '%s'", result["server2"].Command)
	}
}
