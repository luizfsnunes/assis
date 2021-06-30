package assis

import (
	"os"
	"runtime"
	"strings"
)

func GetOSBySystem() OS {
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		return Windows{}
	case "linux":
		return Linux{}
	case "darwin":
		return MaCOSX{}
	default:
		return Linux{}
	}
}

type OS interface {
	ShouldGenerate(string) bool
	WriteFlags() int
}

type Windows struct{}

func (w Windows) ShouldGenerate(op string) bool {
	return op == "REMOVE" || op == "CREATE"
}

func (w Windows) WriteFlags() int {
	return os.O_CREATE
}

type Linux struct{}

func (l Linux) ShouldGenerate(op string) bool {
	return  op == "REMOVE" || op == "WRITE"
}

func (l Linux) WriteFlags() int {
	return os.O_WRONLY | os.O_CREATE
}

type MaCOSX struct {}

func (m MaCOSX) ShouldGenerate(op string) bool{
	return  op == "REMOVE" || op == "WRITE"
}

func (m MaCOSX) WriteFlags() int {
	return os.O_WRONLY | os.O_CREATE
}
