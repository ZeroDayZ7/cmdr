package tree

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zerodayz7/cmdr/internal/utils"
)

func formatJSON(path string, excludeList []string) (string, error) {
	node, err := buildJSONNode(path, excludeList)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildJSONNode(path string, excludeList []string) (Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Node{}, err
	}

	node := Node{Name: info.Name(), Type: "file"}
	if info.IsDir() {
		node.Type = "dir"
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if utils.ShouldExclude(e.Name(), excludeList) {
				continue
			}
			child, _ := buildJSONNode(filepath.Join(path, e.Name()), excludeList)
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}
