package sync

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/go-github/v66/github"
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

func TestConvertPropertyValue(t *testing.T) {
	tests := []struct {
		name           string
		props          []*github.CustomPropertyValue
		failedPropName string
		want           []*github.CustomPropertyValue
	}{
		{
			name: "convert single-select to multi-select",
			props: []*github.CustomPropertyValue{
				{
					PropertyName: "Domain",
					Value:        "Frontend",
				},
				{
					PropertyName: "Other",
					Value:        "unchanged",
				},
			},
			failedPropName: "Domain",
			want: []*github.CustomPropertyValue{
				{
					PropertyName: "Domain",
					Value:        []string{"Frontend"},
				},
				{
					PropertyName: "Other",
					Value:        "unchanged",
				},
			},
		},
		{
			name: "non-string value remains unchanged",
			props: []*github.CustomPropertyValue{
				{
					PropertyName: "Count",
					Value:        42,
				},
			},
			failedPropName: "Count",
			want: []*github.CustomPropertyValue{
				{
					PropertyName: "Count",
					Value:        42,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPropertyValue(tt.props, tt.failedPropName)

			if len(got) != len(tt.want) {
				t.Errorf("convertPropertyValue() returned %d properties, want %d", len(got), len(tt.want))
				return
			}

			for i, prop := range got {
				if prop.PropertyName != tt.want[i].PropertyName {
					t.Errorf("Property %d name = %v, want %v", i, prop.PropertyName, tt.want[i].PropertyName)
				}

				// For the failed property, check if it was converted to []string
				if prop.PropertyName == tt.failedPropName {
					if strVal, ok := tt.props[i].Value.(string); ok {
						// Should be converted to []string
						if arr, ok := prop.Value.([]string); !ok || len(arr) != 1 || arr[0] != strVal {
							t.Errorf("Property %s value = %v, want []string{%v}", prop.PropertyName, prop.Value, strVal)
						}
					}
				} else {
					// Other properties should remain unchanged
					if prop.Value != tt.want[i].Value {
						t.Errorf("Property %s value = %v, want %v", prop.PropertyName, prop.Value, tt.want[i].Value)
					}
				}
			}
		})
	}
}

func TestExtractPropertyName(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		want   string
	}{
		{
			name:   "standard error message",
			errMsg: "422 Property 'Domain' values must be strings",
			want:   "Domain",
		},
		{
			name:   "error message with multiple quotes",
			errMsg: "Property 'Team' has 'invalid' format",
			want:   "Team",
		},
		{
			name:   "error message without property name",
			errMsg: "Invalid request",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPropertyName(tt.errMsg)
			if got != tt.want {
				t.Errorf("extractPropertyName() = %v, want %v", got, tt.want)
			}
		})
	}
}
