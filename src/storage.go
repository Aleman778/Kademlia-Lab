package main

import (
	"sync"
    "time"
)

type Storage struct {
	mappedData map[string]*WrappedData
	mutex sync.RWMutex
}

type WrappedData struct {
	data []byte
	expire int64
}

// const maxExpire = 86400
const maxExpire = 10

func NewStorage() *Storage {
	return &Storage{
        make(map[string]*WrappedData),
        sync.RWMutex{}}
}

func (storage *Storage) Store(hash string, data []byte) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	_, ok := storage.mappedData[hash]
	if ok {
        now := time.Now()
		storage.mappedData[hash].expire = now.Unix() + maxExpire
	} else {
        now := time.Now()
		storage.mappedData[hash] = &WrappedData{data, now.Unix() + maxExpire}
	}
}

func (storage *Storage) Refresh(hash string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	_, ok := storage.mappedData[hash]
	if ok {
        now := time.Now()
		storage.mappedData[hash].expire = now.Unix() + maxExpire
    }
}

func (storage *Storage) Delete(hash string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	delete(storage.mappedData, hash)
}

func (storage *Storage) Load(hash string) ([]byte, bool) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	if v, ok := storage.mappedData[hash]; ok != false {
        now := time.Now()
        sec := now.Unix()
        if (sec >= storage.mappedData[hash].expire) {
            delete(storage.mappedData, hash);
        } else {
            storage.mappedData[hash].expire = sec + maxExpire
            return v.data, true
        }
	}
	return nil, false
}

