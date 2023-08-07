package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	port         = flag.String("port", ":8080", "Define the server port")
	saveFilePath = flag.String("savepath", "data.json", "Define the path to save the key-value store data")
	saveInterval = flag.Duration("saveinterval", 10*time.Second, "Define the interval to automatically save data")
)

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
	if time.Since(store.lastSaved) > *saveInterval {
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
	flag.Parse()

	store := NewKeyValueStore(*saveFilePath)

	err := store.LoadFromFile(*saveFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

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

	log.Printf("Starting server on %s\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
