package fileops

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	setFileAttributes = kernel32.NewProc("SetFileAttributesW")
)

func HideDotDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			HideFile(filepath.Join(dir, file.Name()))
		}
	}
}

func ShowDotDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			ShowFile(filepath.Join(dir, file.Name()))
		}
	}
}

func HideFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
	}
	// 设置隐藏属性
	_ = syscall.SetFileAttributes(ptr, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
	// 设置隐藏属性和系统属性
	//_, _, err = setFileAttributes.Call(uintptr(unsafe.Pointer(ptr)), uintptr(attrs|syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM))
	//if err != nil && err != syscall.Errno(0) {
	//	panic(err)
	//}
}

func ShowFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
	}
	// 移除隐藏属性
	_ = syscall.SetFileAttributes(ptr, attrs&^syscall.FILE_ATTRIBUTE_HIDDEN)
	// 移除隐藏属性和系统属性
	//_, _, err = setFileAttributes.Call(uintptr(unsafe.Pointer(ptr)), uintptr(attrs&^(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM)))
	//if err != nil && err != syscall.Errno(0) {
	//	panic(err)
	//}
}
