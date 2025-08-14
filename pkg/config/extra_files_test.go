package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtraFilesHelper(t *testing.T) {
	t.Run("AddFile", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		helper.AddFile("test.txt", "content")

		files := helper.Build()
		if files["test.txt"] != "content" {
			t.Errorf("expected content, got %s", files["test.txt"])
		}
	})

	t.Run("AddJSON", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		data := map[string]string{"key": "value"}

		err := helper.AddJSON("config.json", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		files := helper.Build()
		if !strings.Contains(files["config.json"], `"key": "value"`) {
			t.Errorf("JSON not properly marshaled: %s", files["config.json"])
		}
	})

	t.Run("AddFileFromPath", func(t *testing.T) {
		// Create a temporary file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := "test file content"

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		helper := NewExtraFilesHelper()
		err = helper.AddFileFromPath("loaded.txt", testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		files := helper.Build()
		if files["loaded.txt"] != testContent {
			t.Errorf("expected %q, got %q", testContent, files["loaded.txt"])
		}
	})

	t.Run("AddFileFromPath_NonExistent", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		err := helper.AddFileFromPath("test.txt", "/non/existent/file.txt")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("AddFileFromReader", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		content := "content from reader"
		reader := strings.NewReader(content)

		err := helper.AddFileFromReader("from-reader.txt", reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		files := helper.Build()
		if files["from-reader.txt"] != content {
			t.Errorf("expected %q, got %q", content, files["from-reader.txt"])
		}
	})

	t.Run("AddFileFromReader_BytesBuffer", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		content := []byte("binary content")
		reader := bytes.NewBuffer(content)

		err := helper.AddFileFromReader("binary.dat", reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		files := helper.Build()
		if files["binary.dat"] != string(content) {
			t.Errorf("expected %q, got %q", string(content), files["binary.dat"])
		}
	})

	t.Run("AddFileFromReader_ErrorCase", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		// Create a reader that always returns an error
		reader := &errorReader{}

		err := helper.AddFileFromReader("error.txt", reader)
		if err == nil {
			t.Error("expected error from failing reader")
		}
		if !strings.Contains(err.Error(), "failed to read content") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("AddFilesFromDirectory", func(t *testing.T) {
		// Create a temporary directory with test files
		tmpDir := t.TempDir()

		// Create test files
		files := map[string]string{
			"file1.txt":  "content 1",
			"file2.yaml": "key: value",
			"file3.json": `{"test": true}`,
		}

		for name, content := range files {
			err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
			if err != nil {
				t.Fatalf("failed to create test file %s: %v", name, err)
			}
		}

		// Create a subdirectory (should be ignored)
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}
		err = os.WriteFile(filepath.Join(subDir, "ignored.txt"), []byte("should not be included"), 0644)
		if err != nil {
			t.Fatalf("failed to create file in subdirectory: %v", err)
		}

		helper := NewExtraFilesHelper()
		err = helper.AddFilesFromDirectory(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result := helper.Build()

		// Verify all files were added
		for name, expectedContent := range files {
			if result[name] != expectedContent {
				t.Errorf("file %s: expected %q, got %q", name, expectedContent, result[name])
			}
		}

		// Verify subdirectory file was not included
		if _, exists := result["ignored.txt"]; exists {
			t.Error("subdirectory file should not be included")
		}

		// Verify count
		if helper.Count() != len(files) {
			t.Errorf("expected %d files, got %d", len(files), helper.Count())
		}
	})

	t.Run("AddFilesFromDirectory_Empty", func(t *testing.T) {
		tmpDir := t.TempDir()

		helper := NewExtraFilesHelper()
		err := helper.AddFilesFromDirectory(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if helper.Count() != 0 {
			t.Errorf("expected 0 files from empty directory, got %d", helper.Count())
		}
	})

	t.Run("AddFilesFromDirectory_NonExistent", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		err := helper.AddFilesFromDirectory("/non/existent/directory")
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
		if !strings.Contains(err.Error(), "failed to read directory") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("AddFilesFromDirectory_UnreadableFile", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("skipping test when running as root")
		}

		tmpDir := t.TempDir()

		// Create a file that we can't read
		unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
		err := os.WriteFile(unreadableFile, []byte("content"), 0000)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		helper := NewExtraFilesHelper()
		err = helper.AddFilesFromDirectory(tmpDir)
		if err == nil {
			t.Error("expected error for unreadable file")
		}
		if !strings.Contains(err.Error(), "failed to read file") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Count", func(t *testing.T) {
		helper := NewExtraFilesHelper()
		helper.AddFile("file1.txt", "content1")
		helper.AddFile("file2.txt", "content2")

		if helper.Count() != 2 {
			t.Errorf("expected 2 files, got %d", helper.Count())
		}
	})

	t.Run("ChainedOperations", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("file content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		helper := NewExtraFilesHelper()

		// Chain multiple operations
		helper.AddFile("inline.txt", "inline content")

		err = helper.AddFileFromPath("from-path.txt", testFile)
		if err != nil {
			t.Fatalf("AddFileFromPath failed: %v", err)
		}

		err = helper.AddFileFromReader("from-reader.txt", strings.NewReader("reader content"))
		if err != nil {
			t.Fatalf("AddFileFromReader failed: %v", err)
		}

		err = helper.AddJSON("config.json", map[string]bool{"enabled": true})
		if err != nil {
			t.Fatalf("AddJSON failed: %v", err)
		}

		// Verify all files are present
		files := helper.Build()
		if len(files) != 4 {
			t.Errorf("expected 4 files, got %d", len(files))
		}

		expectedContents := map[string]string{
			"inline.txt":      "inline content",
			"from-path.txt":   "file content",
			"from-reader.txt": "reader content",
		}

		for name, expected := range expectedContents {
			if files[name] != expected {
				t.Errorf("file %s: expected %q, got %q", name, expected, files[name])
			}
		}

		// Check JSON file contains expected content
		if !strings.Contains(files["config.json"], `"enabled": true`) {
			t.Errorf("JSON file doesn't contain expected content: %s", files["config.json"])
		}
	})
}

// errorReader is a mock io.Reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestEthereumPackageConfigExtraFiles(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		cfg := DefaultValidConfig()
		cfg.ExtraFiles = map[string]string{
			"valid.txt": "content",
			"":          "empty name", // Should fail
		}

		cfg.ApplyDefaults()
		err := cfg.Validate()
		if err == nil {
			t.Error("expected validation error for empty file name")
		}
	})

	t.Run("PathSeparatorValidation", func(t *testing.T) {
		cfg := DefaultValidConfig()
		cfg.ExtraFiles = map[string]string{
			"path/to/file.txt": "content", // Should fail
		}

		cfg.ApplyDefaults()
		err := cfg.Validate()
		if err == nil {
			t.Error("expected validation error for path separator in name")
		}
	})
}

func TestConfigBuilderExtraFiles(t *testing.T) {
	t.Run("WithExtraFile", func(t *testing.T) {
		builder := NewConfigBuilder()
		builder.WithParticipants(DefaultValidConfig().Participants)
		builder.WithExtraFile("test.txt", "content")

		cfg, err := builder.Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ExtraFiles["test.txt"] != "content" {
			t.Error("extra file not set correctly")
		}
	})

	t.Run("WithExtraFiles", func(t *testing.T) {
		files := map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
		}

		builder := NewConfigBuilder()
		builder.WithParticipants(DefaultValidConfig().Participants)
		builder.WithExtraFiles(files)

		cfg, err := builder.Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.ExtraFiles) != 2 {
			t.Errorf("expected 2 files, got %d", len(cfg.ExtraFiles))
		}
	})
}
