package copy

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

// Directory will copy the contents of a directory
func Directory(src string, dest string, numberOfDocs *int) error {
	if err := CreateIfNotExists(dest, os.ModePerm); err != nil {
		return nil
	}

	entries, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)

		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)

		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, os.ModePerm); err != nil {
				return err
			}

			if err := Directory(sourcePath, destPath, numberOfDocs); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := SymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := File(sourcePath, destPath, numberOfDocs); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// File will copy the file
func File(srcFile string, dest string, numberOfDocs *int) error {
	out, err := os.Create(dest)
	defer out.Close()

	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()

	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	if err != nil {
		return err
	}

	*numberOfDocs++

	return nil
}

// Exists will chek if the file exists
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// CreateIfNotExists will create a directory
func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// SymLink will copy sym link files
func SymLink(source, dest string) error {
	link, err := os.Readlink(source)

	if err != nil {
		return err
	}

	return os.Symlink(link, dest)
}

// HomeDir for Windows, Linux, MacOs
func HomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}
