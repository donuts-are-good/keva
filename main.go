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
	saveInterval = flag.Duration("saveinterval", 5*time.Second, "Define the interval to automatically save data")
)

type KeyValueStore struct {
	data         map[string]interface{}
	mutex        sync.RWMutex
	savePath     string
	lastSaved    time.Time
	saveInterval time.Duration
}

func main() {
	flag.Parse()

	store := NewKeyValueStore(*saveFilePath, *saveInterval)

	err := store.LoadFromFile(*saveFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Fatal error during loading: %v", err)
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

func NewKeyValueStore(savePath string, saveInterval time.Duration) *KeyValueStore {
	log.Println("Initializing new key-value store...")
	kvStore := &KeyValueStore{
		data:         make(map[string]interface{}),
		savePath:     savePath,
		saveInterval: saveInterval,
	}
	go kvStore.periodicPersist()
	return kvStore
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

func (store *KeyValueStore) periodicPersist() {
	for {
		time.Sleep(store.saveInterval)
		if time.Since(store.lastSaved) > store.saveInterval {
			store.persist()
		}
	}
}

func (store *KeyValueStore) persist() {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	err := store.SaveToFile(store.savePath)
	if err != nil {
		log.Printf("Error during persistence: %v", err)
	}
	store.lastSaved = time.Now()
}

func (store *KeyValueStore) SaveToFile(filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(store.data)
	if err != nil {
		log.Printf("Error encoding data to file: %v", err)
	}
	return err
}

func (store *KeyValueStore) LoadFromFile(filename string) error {
	log.Printf("Loading from file: %s", filename)
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Println("Initializing datastore...")
		file, createErr := os.Create(filename)
		if createErr != nil {
			log.Printf("Error creating new file: %v", createErr)
			return createErr
		}
		file.WriteString("{}")
		file.Close()
		return nil
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading from file: %v", err)
		return err
	}

	err = json.Unmarshal(data, &store.data)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
	}
	return err
}
