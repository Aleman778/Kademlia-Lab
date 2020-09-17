package d7024e

import (
	"sync"
)

type Storage struct {
	mappedData map[string]*WrappedData
	mutex sync.RWMutex
}

type WrappedData struct {
	data []byte
	expire int
}

const maxExpire = 86400

func NewStorage() *Storage {
	return &Storage{make(map[string]*WrappedData), sync.RWMutex{}}
}

func (storage *Storage) Store(hash string, data []byte) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	_, ok := storage.mappedData[hash]
	if ok {
		storage.mappedData[hash].expire = maxExpire
	} else {
		storage.mappedData[hash] = &WrappedData{data, maxExpire}
	}
}

func (storage *Storage) Delete(hash string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	delete(storage.mappedData, hash)
}

func (storage *Storage) Load(hash string) ([]byte, bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	if v, ok := storage.mappedData[hash]; ok != false {
		return v.data, true
	}
	return nil, false
}
