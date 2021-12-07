package common

import (
	"fmt"
	"os"
	"syscall"
)

type FileLock struct {
	dir string
	f   *os.File
}

func NewFileLock(dir string, create bool) (*FileLock, error) {
	if !FileExists(dir) && create {
		if file, err := os.Create(dir); err != nil {
			return nil, err
		} else {
			return &FileLock{
				dir: dir,
				f:   file,
			}, nil
		}
	}
	return &FileLock{
		dir: dir,
	}, nil
}

func (l *FileLock) Lock() error {
	if l.f == nil {
		f, err := os.Open(l.dir)
		if err != nil {
			return err
		}
		l.f = f
	}
	err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_EX)
	if err != nil {
		return fmt.Errorf("cannot flock directory %s - %s", l.dir, err)
	}
	return nil
}

func (l *FileLock) Unlock() error {
	defer l.f.Close()
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
