package main

import (
	"sync"
    "time"
    "fmt"
)

type RefreshTicker struct {
    ticker *time.Ticker
    forgetCh chan bool
}

type Storage struct {
    refreshStorage map[string]RefreshTicker // NOTE(alexander): needs to be separate from mappedData!
	mappedData map[string]*WrappedData
	mutex sync.Mutex
}

type WrappedData struct {
	data []byte
    expire int64
    expireTimer *time.Timer
}

const maxExpire = 86400
const maxExpireCache = 10

func NewStorage() *Storage {
	return &Storage{
        make(map[string]RefreshTicker),
        make(map[string]*WrappedData),
        sync.Mutex{}}
}

func (storage *Storage) Store(hash string, data []byte, expire int64) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
    v, ok := storage.mappedData[hash];
	if !ok {
        is_cache := false
        if (expire <= 0) { // NOTE(alexander): assumes that caching is done when expire is 0
            fmt.Println("Caching object with hash and data:")
            fmt.Println(hash)
            fmt.Println(data)
            expire = maxExpireCache
            is_cache = true
        }
        
        expireTimer := time.NewTimer(time.Duration(expire)*time.Second)
		storage.mappedData[hash] = &WrappedData{data, expire, expireTimer}
        go func() {
            <-expireTimer.C
            storage.Delete(hash, is_cache)
            v2, ok2 := storage.mappedData[hash]
            fmt.Print("Data object timing out: ")
            fmt.Println(ok2)
            fmt.Print("with object: ")
            fmt.Println(v2)
        }()
	} else {
        if (v.expireTimer != nil) {
            v.expireTimer.Reset(time.Duration(v.expire)*time.Second)
        }
    }
}

func (storage *Storage) RefreshDataPeriodically(hash string, expire int64) *RefreshTicker {
    storage.mutex.Lock()
    defer storage.mutex.Unlock()
    _, ok := storage.refreshStorage[hash]
    if expire <= 5 || ok {
        return nil
    }

    ticker := time.NewTicker(time.Duration(expire - 3)*time.Second)
    forgetCh := make(chan bool)
    refreshTicker := RefreshTicker{ticker, forgetCh}
    storage.refreshStorage[hash] = refreshTicker
    return &refreshTicker
}

func (storage *Storage) StopDataRefresh(hash string) {
    storage.mutex.Lock()
    defer storage.mutex.Unlock()
    t, ok := storage.refreshStorage[hash]
    if ok {
        t.ticker.Stop()
        t.forgetCh <- true
        delete(storage.refreshStorage, hash)
    }
}

func (storage *Storage) Delete(hash string, is_cache bool) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
    v, ok := storage.mappedData[hash];
    if ok && v.expireTimer != nil && !is_cache {
        v.expireTimer.Stop()
        t, ok := storage.refreshStorage[hash]
        if ok {
            t.ticker.Stop()
            t.forgetCh <- true
            delete(storage.refreshStorage, hash)
        }
    }
	delete(storage.mappedData, hash)
}

func (storage *Storage) Load(hash string) ([]byte, bool) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	if v, ok := storage.mappedData[hash]; ok {
        if (v.expireTimer != nil) {
            v.expireTimer.Reset(time.Duration(v.expire)*time.Second)
        }
        return v.data, true
    }
	return nil, false
}
