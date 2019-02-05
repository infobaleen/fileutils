package fileutils

import (
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/infobaleen/errors"
)

func IsNumberedName(name string, prefix, suffix string) (uint64, bool) {
	if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
		var number, err = strconv.ParseUint(name[len(prefix):len(name)-len(suffix)], 10, 64)
		if err != nil {
			return 0, false
		}
		return number, true
	}
	return 0, false
}

func FindNumberedFiles(dir, prefix, suffix string) ([]uint64, []string, error) {
	var contents, err = ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	var numbers []uint64
	var names []string
	for _, info := range contents {
		var name = info.Name()
		if number, ok := IsNumberedName(name, prefix, suffix); ok {
			var idx = sort.Search(len(numbers), func(i int) bool { return number <= numbers[i] })
			if idx < len(numbers) && number == numbers[idx] {
				return nil, nil, errors.Fmt("duplicate file with number %d", number)
			}
			numbers = append(numbers, 0)
			names = append(names, "")
			copy(numbers[idx+1:], numbers[idx:])
			copy(names[idx+1:], names[idx:])
			numbers[idx] = number
			names[idx] = name
		}
	}
	return numbers, names, nil
}

func FindConsecutiveFiles(dir, prefix, suffix string) ([]string, error) {
	var numbers, names, err = FindNumberedFiles(dir, prefix, suffix)
	if err != nil {
		return nil, err
	}
	for idx, number := range numbers {
		if uint64(idx) != number {
			return nil, errors.Fmt("non-consecutive numbers in file names: %d != %d", idx, number)
		}
	}
	return names, nil
}

// FindMaxTimestamp file returns the timestamp and name of the file matching the pattern that has the highest timestamp.
// If no matching files were found, the cause of the returned error is os.ErrNotExist (see github.com/pkg/errors).
func FindMaxTimestampFile(dir, prefix, suffix string) (int64, string, error) {
	var numbers, names, err = FindNumberedFiles(dir, prefix, suffix)
	if err != nil {
		return math.MinInt64, "", err
	}
	if len(numbers) == 0 {
		return math.MinInt64, "", errors.Wrap(os.ErrNotExist, "no files found")
	}
	var lastIdx = len(numbers) - 1
	var lastNumber = numbers[lastIdx]
	if lastNumber > math.MaxInt64 {
		return math.MinInt64, "", errors.Fmt("%d too large for timestamp", lastNumber)
	}
	return int64(lastNumber), names[lastIdx], nil
}
