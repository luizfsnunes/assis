package assis

import (
	"os"
	"path"
)

func GenerateDir(outputFile string) error {
	outputPath := path.Dir(outputFile)
	exists, err := Exists(outputPath)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(outputPath, 0600)
	}
	return nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateTargetFile(output string) (*os.File, error) {
	if err := GenerateDir(output); err != nil {
		return nil, err
	}

	target, err := os.OpenFile(output, os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return target, nil
}
