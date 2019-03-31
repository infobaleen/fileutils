package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ChangeName(oldPath, newName string) (string, error) {
	return Move(oldPath, filepath.Join(filepath.Dir(oldPath), newName))
}

func Move(oldPath, newPath string) (string, error) {
	var err = os.Rename(oldPath, newPath)
	if err != nil {
		return oldPath, err
	}
	return newPath, nil
}

func ExtendName(currentPath, prefix, suffix string) (string, error) {
	var file = filepath.Base(currentPath)
	return ChangeName(currentPath, prefix+file+suffix)
}

func TrimName(currentPath, prefix, suffix string) (string, error) {
	var file = filepath.Base(currentPath)
	if !strings.HasPrefix(file, prefix) {
		return currentPath, fmt.Errorf(`"%s" does not have prefix "%s" and suffix "%s"`, file, prefix, suffix)
	}
	return ChangeName(currentPath, file[len(prefix):len(file)-len(suffix)])
}
