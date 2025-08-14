package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtraFilesHelper provides utilities for working with extra files
type ExtraFilesHelper struct {
	files map[string]string
}

// NewExtraFilesHelper creates a new helper for managing extra files
func NewExtraFilesHelper() *ExtraFilesHelper {
	return &ExtraFilesHelper{
		files: make(map[string]string),
	}
}

// AddFile adds a file with inline content
func (h *ExtraFilesHelper) AddFile(name, content string) *ExtraFilesHelper {
	h.files[name] = content
	return h
}

// AddFileFromPath reads a file from disk and adds it
func (h *ExtraFilesHelper) AddFileFromPath(name, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	h.files[name] = string(content)
	return nil
}

// AddFileFromReader reads content from an io.Reader
func (h *ExtraFilesHelper) AddFileFromReader(name string, r io.Reader) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}
	h.files[name] = string(content)
	return nil
}

// AddJSON marshals an object to JSON and adds it as a file
func (h *ExtraFilesHelper) AddJSON(name string, v interface{}) error {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	h.files[name] = string(content)
	return nil
}

// AddFilesFromDirectory adds all files from a directory
func (h *ExtraFilesHelper) AddFilesFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(dir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			h.files[entry.Name()] = string(content)
		}
	}
	return nil
}

// Build returns the files map for use in NetworkParams
func (h *ExtraFilesHelper) Build() map[string]string {
	return h.files
}

// Count returns the number of files
func (h *ExtraFilesHelper) Count() int {
	return len(h.files)
}
