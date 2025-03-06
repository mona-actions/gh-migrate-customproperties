package sync

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestNewRepositoryProperties(t *testing.T) {
	props := NewRepositoryProperties()
	if props == nil {
		t.Fatal("Expected non-nil RepositoryProperties")
		return
	}
	if props.Repositories == nil {
		t.Error("Expected non-nil Repositories map")
	}
	if len(props.Repositories) != 0 {
		t.Error("Expected empty Repositories map")
	}
}

func TestPrintSyncSummary(t *testing.T) {
	tests := []struct {
		name     string
		stats    *SyncStats
		contains []string
	}{
		{
			name: "successful sync with no failures",
			stats: &SyncStats{
				TotalProcessed:   2,
				SuccessfulFetch:  2,
				SuccessfulCreate: 2,
			},
			contains: []string{
				"Total repositories processed: 2",
				"Successfully fetched: 2",
				"Successfully created: 2",
			},
		},
		{
			name: "sync with failures",
			stats: &SyncStats{
				TotalProcessed:   3,
				SuccessfulFetch:  2,
				SuccessfulCreate: 1,
				FetchFailures:    []string{"repo1"},
				CreateFailures:   []string{"repo2"},
			},
			contains: []string{
				"Total repositories processed: 3",
				"Successfully fetched: 2",
				"Successfully created: 1",
				"repo1",
				"repo2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printSyncSummary(tt.stats)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check if output contains expected strings
			for _, s := range tt.contains {
				if !bytes.Contains(buf.Bytes(), []byte(s)) {
					t.Errorf("Expected output to contain %q", s)
				}
			}
		})
	}
}
