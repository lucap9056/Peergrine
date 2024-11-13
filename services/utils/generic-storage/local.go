package genericstorage

import (
	"container/heap"
	Auth "peergrine/utils/auth"
	GenericHeap "peergrine/utils/generic-heap"
	"sync"
	"time"
)

// LocalStorageManager is a generic in-memory store that supports expiration of data and tokens.
// It uses a heap to manage data and token expiration efficiently. The store is thread-safe and
// periodically checks for expired items, removing them when necessary.
type LocalStorageManager[T base] struct {
	mutex       *sync.RWMutex
	dataHeap    *GenericHeap.GenericHeap[T]
	data        map[string]T
	tokenHeap   *GenericHeap.GenericHeap[Auth.TokenData]
	tokens      map[string]Auth.TokenData
	closeTicker chan struct{}
}

// NewLocalStorageManager creates a new instance of LocalStorageManager.
// It initializes the heaps and starts a ticker to periodically remove expired items.
func NewLocalStorageManager[T base]() *LocalStorageManager[T] {
	dataHeap := GenericHeap.New(func(a T, b T) bool {
		return a.GetExpiresAt() < b.GetExpiresAt()
	})

	tokenHeap := GenericHeap.New(func(a Auth.TokenData, b Auth.TokenData) bool {
		return a.Exp < b.Exp
	})

	storeManager := &LocalStorageManager[T]{
		mutex:       new(sync.RWMutex),
		dataHeap:    dataHeap,
		data:        make(map[string]T),
		tokenHeap:   tokenHeap,
		tokens:      make(map[string]Auth.TokenData),
		closeTicker: make(chan struct{}),
	}

	go storeManager.removeExpiredSignalTicker()
	return storeManager
}

// removeExpiredSignalTicker runs periodically to remove expired signals and tokens
// based on a 1-second interval. It will stop when the store is closed.
func (store *LocalStorageManager[base]) removeExpiredSignalTicker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-store.closeTicker:
			return
		case <-ticker.C:
			store.removeExpiredSignals()
			store.removeExpiredTokens()
		}
	}
}

// removeExpiredSignals removes all expired data from the store.
// It checks the heap for expired items and removes them in a thread-safe manner.
func (store *LocalStorageManager[base]) removeExpiredSignals() {
	if store.dataHeap.Len() > 0 {
		now := time.Now().Unix()

		store.mutex.Lock()
		for store.dataHeap.Len() > 0 {
			data := store.dataHeap.First()
			if data.GetExpiresAt() > now {
				break
			}
			heap.Pop(store.dataHeap)
			delete(store.data, data.GetKey())
		}
		store.mutex.Unlock()
	}
}

// removeExpiredTokens removes expired tokens from the token store.
func (store *LocalStorageManager[any]) removeExpiredTokens() {
	if store.tokenHeap.Len() > 0 {
		now := time.Now().Unix()

		store.mutex.Lock()
		for store.tokenHeap.Len() > 0 {
			token := store.tokenHeap.First()
			if token.Exp > now {
				break
			}
			heap.Pop(store.tokenHeap)
			delete(store.tokens, token.Token)
		}
		store.mutex.Unlock()
	}
}

// Set adds a new data entry to the store, pushing it to the heap for expiration tracking.
func (store *LocalStorageManager[T]) Set(data T) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	heap.Push(store.dataHeap, data)
	store.data[data.GetKey()] = data
}

// Get retrieves data by its key. If the key exists, it returns the data; otherwise, it returns nil.
func (store *LocalStorageManager[any]) Get(key string) *any {
	store.mutex.RLock()
	data, exist := store.data[key]
	store.mutex.RUnlock()

	if exist {
		return &data
	}
	return nil
}

// Exist checks if a given key exists in the data store.
func (store *LocalStorageManager[any]) Exists(key string) bool {
	store.mutex.RLock()
	_, exist := store.data[key]
	store.mutex.RUnlock()
	return exist
}

// Remove deletes an entry from the store by its key.
func (store *LocalStorageManager[any]) Remove(key string) {
	store.mutex.RLock()
	_, exist := store.data[key]
	store.mutex.RUnlock()

	if exist {
		store.mutex.Lock()
		delete(store.data, key)
		store.mutex.Unlock()
	}
}

// SetToken adds a new token entry to the token store.
func (store *LocalStorageManager[any]) SetToken(tokenData Auth.TokenData) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	heap.Push(store.tokenHeap, tokenData)
	store.tokens[tokenData.Token] = tokenData
}

// GetToken retrieves a token by its key. If the key exists, it returns the token; otherwise, it returns nil.
func (store *LocalStorageManager[any]) GetToken(key string) *Auth.TokenData {
	store.mutex.RLock()
	tokenData, exist := store.tokens[key]
	store.mutex.RUnlock()

	if exist {
		return &tokenData
	}
	return nil
}

// RemoveToken deletes a token from the store by its key.
func (store *LocalStorageManager[any]) RemoveToken(key string) {
	store.mutex.RLock()
	_, exist := store.tokens[key]
	store.mutex.RUnlock()

	if exist {
		store.mutex.Lock()
		delete(store.tokens, key)
		store.mutex.Unlock()
	}
}

// Close stops the ticker that removes expired items and closes the store.
func (store *LocalStorageManager[any]) Close() {
	close(store.closeTicker)
}
