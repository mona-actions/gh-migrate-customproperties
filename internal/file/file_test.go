package file

import (
	"os"
	"strings"
	"testing"
)

func TestParseRepositoryFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantRepos   []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid repositories with full URLs",
			content:   "https://github.com/org/repo1\nhttps://github.com/org/repo2\nhttps://github.com/different-org/repo3",
			wantRepos: []string{"repo1", "repo2", "repo3"},
			wantErr:   false,
		},
		{
			name:        "empty file",
			content:     "",
			wantRepos:   nil,
			wantErr:     true,
			errContains: "no repositories found in the list",
		},
		{
			name:        "invalid URL",
			content:     "https://github.com/org/repo1\n:invalid:\nhttps://github.com/org/repo3",
			wantRepos:   nil,
			wantErr:     true,
			errContains: "invalid URI",
		},
		{
			name:      "simple owner/repo format",
			content:   "org/repo1\nother-org/repo2",
			wantRepos: []string{"repo1", "repo2"},
			wantErr:   false,
		},
		{
			name:      "mixed formats",
			content:   "org/repo1\nhttps://github.com/org/repo2\nrepo3",
			wantRepos: []string{"repo1", "repo2", "repo3"},
			wantErr:   false,
		},
		{
			name:      "just repo names",
			content:   "repo1\nrepo2\nrepo3",
			wantRepos: []string{"repo1", "repo2", "repo3"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "repo-list-*.txt")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write test content
			if _, err := tmpfile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpfile.Close()

			// Test the function
			got, err := ParseRepositoryFile(tmpfile.Name())

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q but got %q", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Compare results
			if !equalSlices(got, tt.wantRepos) {
				t.Errorf("ParseRepositoryFile() = %v, want %v", got, tt.wantRepos)
			}
		})
	}

	// Test file not found
	t.Run("file not found", func(t *testing.T) {
		_, err := ParseRepositoryFile("nonexistent-file.txt")
		if err == nil {
			t.Error("Expected error for nonexistent file but got none")
		}
	})
}

// Helper function to compare string slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
