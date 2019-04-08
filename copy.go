package fileutils

import (
	"io"
	"os"
)

func Copy(oldPath, newPath string) error {
	var src, err = os.Open(oldPath)
	if err != nil {
		return err
	}
	defer src.Close()

	var dst *File
	dst, err = CreateFileTmp(newPath)
	if err != nil {
		return err
	}
	defer dst.RemoveIfTmp()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return dst.Finalize()
}
