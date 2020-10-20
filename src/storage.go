package main

import (
	"sync"
    "time"
    "fmt"
)

type RefreshTicker struct {
    ticker *time.Ticker
    refreshCh chan int64 // expire time
    forgetCh chan bool
}

type Storage struct {
    refreshStorage map[string]RefreshTicker // NOTE(alexander): needs to be separate from mappedData!
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
        make(map[string]RefreshTicker),
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

func (storage *Storage) Refresh(hash string, expire int64) {
    storage.mutex.Lock()
    defer storage.mutex.Unlock()
    storage.RefreshExpireTimer(hash, expire)
}

// NOTE(alexander): not thread safe, guard function call with mutex!
func (storage *Storage) RefreshExpireTimer(hash string, expire int64)  {
    if (expire < 0) {
        expire = maxExpire
    }

    data, ok := storage.mappedData[hash]
    if !ok {
        return
    }

    fmt.Printf("Refresh expire timeer, TTL = %d seconds\n", expire)
    if data.expireTimer == nil {
        data.expireTimer = time.NewTimer(time.Duration(expire)*time.Second)
        go func() {
            <-data.expireTimer.C
            storage.Delete(hash)
        }()
    } else {
        data.expireTimer.Reset(time.Duration(expire)*time.Second)
    }
}

func (storage *Storage) RefreshDataPeriodically(hash string, expire int64, storing bool) *RefreshTicker {
    storage.mutex.Lock()
    defer storage.mutex.Unlock()
    t, ok := storage.refreshStorage[hash]
    if expire <= 0 {
        return nil
    }

    if ok {
        fmt.Printf("Resetting ticker, refresh storage every %d seconds\n", expire - 3)
        t.refreshCh <- expire // NOTE(alexander): Refresh for safety might reset before new tick.
        t.ticker.Reset(time.Duration(expire - 3)*time.Second)
        return nil
    }

    if !storing {
        return nil
    }

    fmt.Printf("Starting ticker, refresh storage every %d seconds\n", expire - 3)
    ticker := time.NewTicker(time.Duration(expire - 3)*time.Second)
    refreshCh := make(chan int64)
    forgetCh := make(chan bool)
    refreshTicker := RefreshTicker{ticker, refreshCh, forgetCh}
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

func (storage *Storage) Delete(hash string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
    data, ok := storage.mappedData[hash];
    if ok && data.expireTimer != nil {
        data.expireTimer.Stop()
        t, ok := storage.refreshStorage[hash]
        if ok {
            t.ticker.Stop()
            t.forgetCh <- true
            delete(storage.refreshStorage, hash)
        }
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
