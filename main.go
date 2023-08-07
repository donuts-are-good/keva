package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type KeyValueStore struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		data: make(map[string]interface{}),
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
}

func (store *KeyValueStore) Delete(key string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.data, key)
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
	if err := encoder.Encode(store.data); err != nil {
		return err
	}

	return nil
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

	if err := json.Unmarshal(data, &store.data); err != nil {
		return err
	}

	return nil
}

func main() {
	store := NewKeyValueStore()

	// Example usage
	store.Set("key1", "value1")
	store.Set("key2", "value2")

	value, ok := store.Get("key1")
	if ok {
		fmt.Println("Value:", value)
	} else {
		fmt.Println("Key not found")
	}

	store.Delete("key2")

	err := store.SaveToFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	err = store.LoadFromFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	// Start HTTP server to expose the key-value store
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value, ok := store.Get(key)
		if ok {
			fmt.Fprintf(w, "Value: %v", value)
		} else {
			fmt.Fprintf(w, "Key not found")
		}
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		store.Set(key, value)
		fmt.Fprintf(w, "Key-Value set successfully")
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		store.Delete(key)
		fmt.Fprintf(w, "Key deleted successfully")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
