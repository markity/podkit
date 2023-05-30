package tools

import (
	"errors"
	"syscall"
)

// 不线程安全
type FlockManager struct {
	init_already bool
	locked       bool
	fd           int
}

func (fm *FlockManager) Init(path string) error {
	if fm.init_already {
		return errors.New("FlockManager had been initialized already")
	}

	fd, err := syscall.Open(path, syscall.O_RDWR, 0)
	if err != nil {
		return err
	}

	fm.fd = fd
	fm.init_already = true
	fm.locked = false

	return nil
}

func (fm *FlockManager) Release() error {
	if !fm.init_already {
		return errors.New("FlockManager haven't been initialized yet")
	}

	var err error
	if fm.locked {
		err = syscall.Flock(fm.fd, syscall.LOCK_UN)
	}

	syscall.Close(fm.fd)

	fm.init_already = false

	return err
}

func (fm *FlockManager) TryLock() (bool, error) {
	if !fm.init_already {
		return false, errors.New("FlockManager haven't been initialized yet")
	}

	if fm.locked {
		return false, errors.New("trying lock a locked lock")
	}

	err := syscall.Flock(fm.fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EAGAIN {
			return false, nil
		}
		return false, err
	}

	fm.locked = true
	return true, nil
}

func (fm *FlockManager) Lock() error {
	if !fm.init_already {
		return errors.New("FlockManager haven't been initialized yet")
	}

	if fm.locked {
		return errors.New("trying lock a locked lock")
	}
	err := syscall.Flock(fm.fd, syscall.LOCK_EX)
	if err != nil {
		return err
	}

	fm.locked = true
	return nil
}

func (fm *FlockManager) Unlock() error {
	if !fm.init_already {
		return errors.New("FlockManager haven't been initialized yet")
	}

	if !fm.locked {
		return errors.New("not locked yet")
	}
	err := syscall.Flock(fm.fd, syscall.LOCK_UN)
	if err != nil {
		return err
	}

	fm.locked = false

	return nil
}
