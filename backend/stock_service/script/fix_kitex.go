package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := "kitex_gen"

	// 1. General cleanup
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			s := string(content)

			lines := strings.Split(s, "\n")
			var newLines []string
			for _, line := range lines {
				if strings.Contains(line, "code.byted.org/kitex/apache_monitor") {
					continue
				}
				if strings.Contains(line, "apache_warning.WarningApache") {
					continue
				}
				// Clean up internal imports if they remain
				if strings.Contains(line, "byted") {
					continue
				}

				// Replace imports
				line = strings.Replace(line, "code.byted.org/kite/kitex/client", "github.com/cloudwego/kitex/client", -1)
				line = strings.Replace(line, "code.byted.org/kite/kitex/server", "github.com/cloudwego/kitex/server", -1)

				newLines = append(newLines, line)
			}

			newContent := strings.Join(newLines, "\n")
			os.WriteFile(path, []byte(newContent), 0644)
		}
		return nil
	})
}
