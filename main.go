package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const defaultSaveInterval = 10 * time.Second

type KeyValueStore struct {
	data      map[string]interface{}
	mutex     sync.RWMutex
	savePath  string
	lastSaved time.Time
}

func NewKeyValueStore(savePath string) *KeyValueStore {
	return &KeyValueStore{
		data:     make(map[string]interface{}),
		savePath: savePath,
	}
}

func (store *KeyValueStore) Get(key string) (interface{}, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	value, ok := store.data[key]
	return value, ok
}

func (store *KeyValueStore) Set(key string, value interface{}) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.data[key] = value
	store.checkAndPersist()
}

func (store *KeyValueStore) Delete(key string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.data, key)
	store.checkAndPersist()
}

func (store *KeyValueStore) checkAndPersist() {
	if time.Since(store.lastSaved) > defaultSaveInterval {
		store.SaveToFile(store.savePath)
		store.lastSaved = time.Now()
	}
}

func (store *KeyValueStore) SaveToFile(filename string) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(store.data)
}

func (store *KeyValueStore) LoadFromFile(filename string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &store.data)
}

func main() {
	savePath := "data.json"
	store := NewKeyValueStore(savePath)

	// Load existing data
	err := store.LoadFromFile(savePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Start HTTP server to expose the key-value store
	http.HandleFunc("/store/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/store/"):]
		switch r.Method {
		case http.MethodGet:
			value, ok := store.Get(key)
			if !ok {
				http.Error(w, "Key not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%v", value)

		case http.MethodPost:
			value := r.FormValue("value")
			if value == "" {
				http.Error(w, "No value provided", http.StatusBadRequest)
				return
			}
			store.Set(key, value)
			fmt.Fprintf(w, "Key-Value set successfully")

		case http.MethodDelete:
			store.Delete(key)
			fmt.Fprintf(w, "Key deleted successfully")

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Healthy")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
