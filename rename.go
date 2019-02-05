package fileutils

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func ChangeName(oldPath, newName string) (string, error) {
	var newPath = path.Join(path.Dir(oldPath), newName)
	var err = os.Rename(oldPath, newPath)
	if err != nil {
		return oldPath, err
	}
	return newPath, nil
}

func ExtendName(currentPath, prefix, suffix string) (string, error) {
	var file = path.Base(currentPath)
	return ChangeName(currentPath, prefix+file+suffix)
}

func TrimName(currentPath, prefix, suffix string) (string, error) {
	var file = path.Base(currentPath)
	if !strings.HasPrefix(file, prefix) {
		return currentPath, fmt.Errorf(`"%s" does not have prefix "%s" and suffix "%s"`, file, prefix, suffix)
	}
	return ChangeName(currentPath, file[len(prefix):len(file)-len(suffix)])
}
