package main

import (
	"sync"
    "time"
    "fmt"
)

type Storage struct {
	mappedData map[string]*WrappedData
	mutex sync.Mutex
}

type WrappedData struct {
	data []byte
    expireTimer *time.Timer
}

const maxExpire = 86400

func NewStorage() *Storage {
	return &Storage{
        make(map[string]*WrappedData),
        sync.Mutex{}}
}

func (storage *Storage) Store(hash string, data []byte, expire int64) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	_, ok := storage.mappedData[hash]
	if !ok {
		storage.mappedData[hash] = &WrappedData{data, nil}
	}
    storage.RefreshExpireTimer(hash, expire)
}

// NOTE(alexander): not thread safe, guard function call with mutex!
func (storage *Storage) RefreshExpireTimer(hash string, expire int64)  {
    if (expire < 0) {
        expire = maxExpire
    }

    data, _ := storage.mappedData[hash]
    if data.expireTimer == nil {
        data.expireTimer = time.NewTimer(time.Duration(expire)*time.Second)
        go func() {
            <-data.expireTimer.C
            storage.Delete(hash);
        }()
    } else {
        data.expireTimer.Reset(time.Duration(expire)*time.Second)
    }
}

func (storage *Storage) Delete(hash string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
    data, ok := storage.mappedData[hash];
    if ok && data.expireTimer == nil {
        data.expireTimer.Stop()
    }
	delete(storage.mappedData, hash)
}

func (storage *Storage) Load(hash string, expire int64) ([]byte, bool) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	if v, ok := storage.mappedData[hash]; ok != false {
        storage.RefreshExpireTimer(hash, expire)
        return v.data, true
    }
	return nil, false
}
