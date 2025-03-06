package file

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func ParseRepositoryFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var repos []string
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if line == "" {
			continue
		}

		var repoName string
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
			return nil, fmt.Errorf("invalid URI on line %d: %s", lineCount, line)
		}

		if strings.Contains(line, "/") {
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				u, err := url.Parse(line)
				if err != nil {
					return nil, fmt.Errorf("invalid URI on line %d: %v", lineCount, err)
				}
				parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
				if len(parts) > 1 {
					repoName = parts[len(parts)-1]
				} else {
					return nil, fmt.Errorf("invalid URI on line %d: missing repository name", lineCount)
				}
			} else {
				// Handle simple owner/repo format
				parts := strings.Split(line, "/")
				if len(parts) > 1 {
					repoName = parts[len(parts)-1]
				} else {
					return nil, fmt.Errorf("invalid repository format on line %d: must be owner/repo", lineCount)
				}
			}
		} else {
			// If no slash, assume it's just a repo name
			repoName = line
		}

		if repoName != "" {
			repos = append(repos, repoName)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in the list")
	}

	return repos, nil
}
