package fileutils

import (
	"fmt"
	"github.com/infobaleen/errors"
	"path"
	"reflect"
)

const (
	TagKeyBinaryFile = "binary-file"
	TagKeyJsonFile   = "json-file"
)

func PopulateTaggedStruct(dir string, p interface{}) error {
	var val = toValue(p)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.Fmt("expected pointer to struct")
	}
	val = val.Elem()

	for i := 0; i < val.NumField(); i++ {
		if tag := val.Type().Field(i).Tag.Get(TagKeyBinaryFile); tag != "" {
			if err := ReadBinaryFile(path.Join(dir, tag), getAddrInterface(val.Field(i))); err != nil {
				return err
			}
		} else if tag := val.Type().Field(i).Tag.Get(TagKeyJsonFile); tag != "" {
			if err := ReadJsonFile(path.Join(dir, tag), getAddrInterface(val.Field(i))); err != nil {
				return err
			}
		}
	}
	return nil
}

func IterateTaggedStruct(v interface{}, f func(fileType, fileName string, field reflect.Value) error) error {
	var val = recursiveIndirect(toValue(v))
	if val.Kind() != reflect.Struct {
		return errors.Fmt("expected struct or pointer to struct")
	}
	for i := 0; i < val.NumField(); i++ {
		if tag := val.Type().Field(i).Tag.Get(TagKeyBinaryFile); tag != "" {
			if err := f(TagKeyBinaryFile, tag, val.Field(i)); err != nil {
				return err
			}
		} else if tag := val.Type().Field(i).Tag.Get(TagKeyJsonFile); tag != "" {
			if err := f(TagKeyJsonFile, tag, val.Field(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func WriteTaggedStructFiles(dir string, v interface{}) error {
	return IterateTaggedStruct(v, func(fileType, fileName string, field reflect.Value) error {
		var path = path.Join(dir, fileName)
		switch fileType {
		case TagKeyBinaryFile:
			return WriteBinaryFile(path, field)
		case TagKeyJsonFile:
			return WriteJsonFile(path, field)
		default:
			return fmt.Errorf("unknown file type %q", fileType)
		}
	})
}

func toValue(v interface{}) reflect.Value {
	var value, isValue = v.(reflect.Value)
	if !isValue {
		value = reflect.ValueOf(v)
	}
	return value
}

func recursiveIndirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func getInterface(v reflect.Value) interface{} {
	v = recursiveIndirect(v)
	if v.CanAddr() {
		return v.Addr().Interface()
	}
	return v.Interface()
}

func getAddrInterface(v reflect.Value) interface{} {
	return recursiveIndirect(v).Addr().Interface()
}
