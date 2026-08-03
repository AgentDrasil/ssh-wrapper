package files

import (
	"errors"
	"os"
	"syscall"
)

var (
	ErrMissingFile   = errors.New("missing critical file")
	ErrNotOwnedByUid = errors.New("file must be owned by specified uid")
	ErrInsecurePerms = errors.New("file has insecure permissions")
	ErrIsSymlink     = errors.New("file cannot be a symbolic link")
)

func VerifySecurity(path string, expectedUid uint32, expectedMode os.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		return ErrMissingFile
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return ErrIsSymlink
	}

	stat := info.Sys().(*syscall.Stat_t)
	if stat.Uid != expectedUid {
		return ErrNotOwnedByUid
	}

	if info.Mode().Perm() != expectedMode {
		return ErrInsecurePerms
	}

	return nil
}

func VerifyDirectorySecurity(dirPath string, expectedUid uint32, expectedDirMode os.FileMode) error {
	info, err := os.Lstat(dirPath)
	if err != nil {
		return ErrMissingFile
	}

	if !info.IsDir() {
		return errors.New("path is not a directory")
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return ErrIsSymlink
	}

	stat := info.Sys().(*syscall.Stat_t)
	if stat.Uid != expectedUid {
		return ErrNotOwnedByUid
	}

	if info.Mode().Perm() != expectedDirMode {
		return ErrInsecurePerms
	}

	return nil
}
