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
	for scanner.Scan() {
		u, err := url.Parse(scanner.Text())
		if err != nil {
			return nil, err
		}
		repo := strings.TrimPrefix(u.Path, "/") // format to owner/reponame
		repos = append(repos, repo)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in the list")
	}

	return repos, nil
}
