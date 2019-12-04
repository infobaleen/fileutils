package fileutils

import "os"

func Exists(path string) (bool, error) {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
