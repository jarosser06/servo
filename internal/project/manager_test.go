package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestIsProject(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "servo-project-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Should return false when no .servo directory exists
	if manager.IsProject() {
		t.Error("Expected IsProject() to return false when no .servo directory exists")
	}

	// Create .servo directory and project.yaml file
	servoDir := filepath.Join(tmpDir, ".servo")
	err = os.MkdirAll(servoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .servo directory: %v", err)
	}

	// Create a minimal project.yaml file
	projectFile := filepath.Join(servoDir, "project.yaml")
	err = os.WriteFile(projectFile, []byte("name: test\ncreated_at: 2024-01-01T00:00:00Z\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create project.yaml: %v", err)
	}

	// Should return true when .servo/project.yaml exists
	if !manager.IsProject() {
		t.Error("Expected IsProject() to return true when .servo/project.yaml exists")
	}
}

func TestInit(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "servo-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Test successful initialization
	project, err := manager.Init("default", []string{"vscode", "cursor"})
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if len(project.Clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(project.Clients))
	}

	// Verify .servo directory structure was created
	servoDir := filepath.Join(tmpDir, ".servo")
	if _, err := os.Stat(servoDir); os.IsNotExist(err) {
		t.Error("Expected .servo directory to be created")
	}

	// Verify session-based directory structure was created
	sessionDirs := []string{
		"config", // Project-level config
		"sessions",
		"sessions/default",
		"sessions/default/manifests",
		"sessions/default/config", // Session-level config
		"sessions/default/volumes",
	}
	for _, subdir := range sessionDirs {
		dir := filepath.Join(servoDir, subdir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected %s directory to be created", subdir)
		}
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(servoDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error("Expected .gitignore file to be created")
	}

	// Verify project.yaml was created and contains correct data
	projectFile := filepath.Join(servoDir, "project.yaml")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		t.Error("Expected project.yaml file to be created")
	}

	// Test initialization when project already exists
	_, err = manager.Init("default", []string{})
	if err == nil {
		t.Error("Expected error when initializing project in directory that already has one")
	}
}

func TestInit_Simple(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "servo-simple-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Test simple initialization
	project, err := manager.Init("default", []string{})
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	if project.DefaultSession != "default" {
		t.Errorf("Expected default session 'default', got '%s'", project.DefaultSession)
	}
}

func TestGet(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-get-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Test Get() when no project exists
	_, err = manager.Get()
	if err == nil {
		t.Error("Expected error when getting project in directory without one")
	}

	// Initialize a project first
	createdProject, err := manager.Init("default", []string{"vscode"})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Test Get() when project exists
	project, err := manager.Get()
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if project.DefaultSession != createdProject.DefaultSession {
		t.Errorf("Expected default session '%s', got '%s'", createdProject.DefaultSession, project.DefaultSession)
	}

	if len(project.Clients) != len(createdProject.Clients) {
		t.Errorf("Expected %d clients, got %d", len(createdProject.Clients), len(project.Clients))
	}
}

func TestSave(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-save-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Initialize a project
	project, err := manager.Init("default", []string{"vscode"})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Modify the project
	project.Clients = append(project.Clients, "cursor")

	// Save the modified project
	err = manager.Save(project)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify changes were saved
	savedProject, err := manager.Get()
	if err != nil {
		t.Fatalf("Failed to get saved project: %v", err)
	}

	if len(savedProject.Clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(savedProject.Clients))
	}
}

func TestDelete(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-delete-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Test Delete() when no project exists
	err = manager.Delete()
	if err == nil {
		t.Error("Expected error when deleting project in directory without one")
	}

	// Initialize a project
	_, err = manager.Init("default", []string{})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Verify project exists
	if !manager.IsProject() {
		t.Error("Expected project to exist before deletion")
	}

	// Delete the project
	err = manager.Delete()
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Verify project no longer exists
	if manager.IsProject() {
		t.Error("Expected project to be deleted")
	}

	// Verify .servo directory was removed
	servoDir := filepath.Join(tmpDir, ".servo")
	if _, err := os.Stat(servoDir); !os.IsNotExist(err) {
		t.Error("Expected .servo directory to be removed")
	}
}

func TestAddClient(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-addclient-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Initialize a project with one client
	_, err = manager.Init("default", []string{"vscode"})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Add a new client
	err = manager.AddClient("cursor")
	if err != nil {
		t.Fatalf("AddClient() returned error: %v", err)
	}

	// Verify client was added
	project, err := manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(project.Clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(project.Clients))
	}

	// Verify both clients exist
	hasVscode := false
	hasCursor := false
	for _, client := range project.Clients {
		if client == "vscode" {
			hasVscode = true
		}
		if client == "cursor" {
			hasCursor = true
		}
	}

	if !hasVscode || !hasCursor {
		t.Error("Expected both vscode and cursor clients to exist")
	}

	// Test adding duplicate client
	err = manager.AddClient("cursor")
	if err == nil {
		t.Error("Expected error when adding duplicate client")
	}
}

func TestRemoveClient(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-removeclient-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Initialize a project with two clients
	_, err = manager.Init("default", []string{"vscode", "cursor"})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Remove a client
	err = manager.RemoveClient("cursor")
	if err != nil {
		t.Fatalf("RemoveClient() returned error: %v", err)
	}

	// Verify client was removed
	project, err := manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(project.Clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(project.Clients))
	}

	if project.Clients[0] != "vscode" {
		t.Errorf("Expected remaining client to be 'vscode', got '%s'", project.Clients[0])
	}

	// Test removing non-existent client
	err = manager.RemoveClient("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent client")
	}
}

func TestSessionMethods(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-session-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Initialize a project
	_, err = manager.Init("default", []string{})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Set session
	err = manager.SetActiveSession("data-science")
	if err != nil {
		t.Fatalf("SetSession() returned error: %v", err)
	}

	// Verify session was set
	project, err := manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.ActiveSession != "data-science" {
		t.Errorf("Expected session 'data-science', got '%s'", project.ActiveSession)
	}

	// Clear session
	err = manager.ClearActiveSession()
	if err != nil {
		t.Fatalf("ClearSession() returned error: %v", err)
	}

	// Verify session was cleared
	project, err = manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.ActiveSession != "" {
		t.Errorf("Expected empty session, got '%s'", project.ActiveSession)
	}
}

func TestGetProjectPath(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-path-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Test GetProjectPath when no project exists
	_, err = manager.GetProjectPath()
	if err == nil {
		t.Error("Expected error when getting project path in directory without project")
	}

	// Initialize a project
	_, err = manager.Init("default", []string{})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Get project path
	projectPath, err := manager.GetProjectPath()
	if err != nil {
		t.Fatalf("GetProjectPath() returned error: %v", err)
	}

	// Should return current directory - resolve both paths to handle symlinks on macOS
	expectedPath, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		expectedPath = tmpDir // fallback if EvalSymlinks fails
	}
	actualPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		actualPath = projectPath // fallback if EvalSymlinks fails
	}

	if actualPath != expectedPath {
		t.Errorf("Expected project path '%s', got '%s'", expectedPath, actualPath)
	}
}

func TestGetServoDir(t *testing.T) {
	manager := NewManager()

	servoDir := manager.GetServoDir()
	if servoDir != ".servo" {
		t.Errorf("Expected '.servo', got '%s'", servoDir)
	}
}

func TestRequiredSecrets(t *testing.T) {
	// Setup temporary directory with a project
	tmpDir, err := os.MkdirTemp("", "servo-secrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager := NewManager()

	// Initialize a project
	_, err = manager.Init("default", []string{})
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Add required secret
	err = manager.AddRequiredSecret("api_key", "API key for service")
	if err != nil {
		t.Fatalf("AddRequiredSecret() returned error: %v", err)
	}

	// Verify secret was added
	project, err := manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(project.RequiredSecrets) != 1 {
		t.Errorf("Expected 1 required secret, got %d", len(project.RequiredSecrets))
	}

	secret := project.RequiredSecrets[0]
	if secret.Name != "api_key" {
		t.Errorf("Expected secret name 'api_key', got '%s'", secret.Name)
	}

	if secret.Description != "API key for service" {
		t.Errorf("Expected description 'API key for service', got '%s'", secret.Description)
	}

	// Test adding duplicate secret
	err = manager.AddRequiredSecret("api_key", "Different description")
	if err == nil {
		t.Error("Expected error when adding duplicate secret")
	}

	// Add another secret
	err = manager.AddRequiredSecret("db_url", "Database connection URL")
	if err != nil {
		t.Fatalf("AddRequiredSecret() returned error: %v", err)
	}

	// Verify two secrets exist
	project, err = manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(project.RequiredSecrets) != 2 {
		t.Errorf("Expected 2 required secrets, got %d", len(project.RequiredSecrets))
	}

	// Remove a secret
	err = manager.RemoveRequiredSecret("api_key")
	if err != nil {
		t.Fatalf("RemoveRequiredSecret() returned error: %v", err)
	}

	// Verify secret was removed
	project, err = manager.Get()
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if len(project.RequiredSecrets) != 1 {
		t.Errorf("Expected 1 required secret after removal, got %d", len(project.RequiredSecrets))
	}

	if project.RequiredSecrets[0].Name != "db_url" {
		t.Errorf("Expected remaining secret to be 'db_url', got '%s'", project.RequiredSecrets[0].Name)
	}

	// Test removing non-existent secret
	err = manager.RemoveRequiredSecret("non_existent")
	if err == nil {
		t.Error("Expected error when removing non-existent secret")
	}

	// Test GetMissingSecrets
	missingSecrets, err := manager.GetMissingSecrets()
	if err != nil {
		t.Fatalf("GetMissingSecrets() returned error: %v", err)
	}

	// Should return all required secrets as missing (placeholder implementation)
	if len(missingSecrets) != 1 {
		t.Errorf("Expected 1 missing secret, got %d", len(missingSecrets))
	}
}
