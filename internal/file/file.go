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
		line := strings.TrimSpace(scanner.Text())
		lineCount++
		if line == "" {
			continue
		}

		var fullRepo string
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			// Handle full GitHub URLs
			u, err := url.Parse(line)
			if err != nil {
				return nil, fmt.Errorf("invalid URI on line %d: %v", lineCount, err)
			}
			parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid repository format on line %d: URL must be in the format 'https://github.com/owner/repo'", lineCount)
			}
			fullRepo = fmt.Sprintf("%s/%s", parts[0], parts[1])
		} else if strings.Contains(line, "/") {
			// Handle owner/repo format
			parts := strings.Split(line, "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid repository format on line %d: must be in the format 'owner/repo'", lineCount)
			}
			fullRepo = line
		} else {
			return nil, fmt.Errorf("invalid repository format on line %d: must be in the format 'owner/repo' or full GitHub URL", lineCount)
		}

		repos = append(repos, fullRepo)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in the list")
	}

	return repos, nil
}
