package fileutils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"

	"github.com/infobaleen/errors"
)

func Untar(dir string, r io.Reader) error {
	var err = os.MkdirAll(dir, 0777)
	err = errors.WithTrace(err)
	if err != nil {
		return err
	}
	var tr = tar.NewReader(r)
	for {
		var header, err = tr.Next()
		err = errors.WithTrace(err)
		if errors.Cause(err) == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.Mkdir(path.Join(dir, header.Name), 0777)
			err = errors.WithTrace(err)
		case tar.TypeReg:
			err = func() error {
				var f, err = os.OpenFile(path.Join(dir, header.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
				err = errors.WithTrace(err)
				if err != nil {
					return err
				}
				defer f.Close()
				_, err = io.Copy(f, tr)
				err = errors.WithTrace(err)
				if err != nil {
					return err
				}
				return errors.WithTrace(f.Close())
			}()
		default:
			err = fmt.Errorf("unexpected type flag: %d", header.Typeflag)
		}
		if err != nil {
			return err
		}
	}
}

type TarWriter tar.Writer

func NewTarWriter(w io.Writer) *TarWriter {
	return (*TarWriter)(tar.NewWriter(w))
}

func (tw *TarWriter) Close() error {
	return (*tar.Writer)(tw).Close()
}

func (tw *TarWriter) AddDir(dir string) error {
	return (*tar.Writer)(tw).WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     dir,
		Mode:     0777,
	})
}

func (tw *TarWriter) AddFileFunc(file string, size int64, f func(contentWriter io.Writer) error) error {
	var err = (*tar.Writer)(tw).WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     file,
		Mode:     0666,
		Size:     size,
	})
	if err != nil {
		return err
	}
	err = f((*tar.Writer)(tw))
	if err != nil {
		return err
	}
	return (*tar.Writer)(tw).Flush()
}

func (tw *TarWriter) AddFileBytes(file string, content []byte) error {
	return tw.AddFileFunc(file, int64(len(content)), func(contentWriter io.Writer) error {
		var _, err = contentWriter.Write(content)
		return err
	})
}

func (tw *TarWriter) AddFileJson(file string, content ...interface{}) error {
	var buffer bytes.Buffer
	var err = WriteJson(&buffer, content...)
	if err != nil {
		return err
	}
	return tw.AddFileBytes(file, buffer.Bytes())
}

func (tw *TarWriter) AddFileBinary(file string, content interface{}) error {
	var v = recursiveIndirect(toValue(content))
	return tw.AddFileFunc(file, int64(sizeBinary(v)), func(contentWriter io.Writer) error {
		return writeBinary(contentWriter, v)
	})
}

func (tw *TarWriter) AddFile(file string, size int64, content io.Reader) error {
	return tw.AddFileFunc(file, size, func(contentWriter io.Writer) error {
		var _, err = io.Copy(contentWriter, content)
		return err
	})
}

func (tw *TarWriter) AddTaggedStruct(archivePrefix string, v interface{}) error {
	return IterateTaggedStruct(v, func(fileType, fileName string, field reflect.Value) error {
		var path = path.Join(archivePrefix, fileName)
		switch fileType {
		case TagKeyBinaryFile:
			return tw.AddFileBinary(path, field)
		case TagKeyJsonFile:
			return tw.AddFileJson(path, field)
		default:
			return fmt.Errorf("unknown file type %q", fileType)
		}
	})
}

// AddPath reads a file or directory from a path and adds its content to the archive, relative to the specified prefix.
// The caller must ensure any directory specified in the prefix was previously added to the archive:
//	tw.AddDir("top")
//	tw.Add("top","path/to/file")
// If path points to a directory only its contents are added. To add both the directory and the contents do:
//	tw.AddDir("top")
//	tw.AddDir("top/dir")
//	tw.Add("top/dir","path/to/dir")
func (tw *TarWriter) AddPath(archivePrefix, diskPath string) error {
	var info, err = os.Stat(diskPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return tw.addDir(archivePrefix, diskPath)
	} else if info.Mode().IsRegular() {
		return tw.addFile(path.Join(archivePrefix, info.Name()), diskPath, info.Size())
	}
	return errors.Fmt("\"%s\" is neither directory nor regular file", diskPath)
}

func (tw *TarWriter) addDir(prefix, dir string) error {
	var infos, err = ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		var name = info.Name()
		var diskPath = path.Join(dir, name)
		var archivePath = path.Join(prefix, name)
		if info.IsDir() {
			err = tw.AddDir(archivePath)
			if err != nil {
				return err
			}
			err = tw.addDir(archivePath, diskPath)
		} else if info.Mode().IsRegular() {
			err = tw.addFile(archivePath, diskPath, info.Size())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (tw *TarWriter) addFile(archivePath, diskPath string, size int64) error {
	var f, err = os.Open(diskPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return tw.AddFile(archivePath, size, f)
}
