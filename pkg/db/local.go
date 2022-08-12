package db

import (
	"errors"
	"sync"
)

type Empty struct{}

var empty Empty

const (
	ErrNotElemFound int = iota
)

var errMaps map[int]error = map[int]error{
	ErrNotElemFound: errors.New("the given element not found in the localStorage"),
}

type LocalStorage struct {
	mu      sync.Mutex
	storage []string
	record  map[string]Empty
}

func NewLocalStorage() (*LocalStorage, error) {
	return &LocalStorage{record: make(map[string]Empty)}, nil
}

func (ls *LocalStorage) copy() []string {
	var newslice []string = make([]string, len(ls.storage))
	copy(newslice, ls.storage)
	return newslice
}

func (ls *LocalStorage) Append(elem ...string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for _, item := range elem {
		if _, exists := ls.record[item]; exists {
			return nil
		}
		ls.storage = append(ls.storage, item)
		ls.record[item] = empty
	}

	return nil
}

func (ls *LocalStorage) Delete(elem string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	var match int
	for index, cur := range ls.storage {
		if cur == elem {
			match = index
			break
		}
	}

	if match == len(ls.storage)-1 && ls.storage[match] != elem {
		return errMaps[ErrNotElemFound]
	}

	delete(ls.record, ls.storage[match])
	ls.storage = append(ls.storage[:match], ls.storage[match+1:]...)

	return nil
}

func (ls *LocalStorage) List() ([]string, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	return ls.copy(), nil
}
