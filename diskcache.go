package diskcache

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// Cache is a disk cache.
// It stores entries in a directory on disk.
type Cache struct {
	dir string
}

// Data is a cache entry.
// It contains a key, a value, and an expiry time.
// Because the disk cache hashes the key for a filename, the key is stored in the entry.
// The hash ensures that the filename is valid and unique.
type Data struct {
	Expiry time.Time
	Key    string
	Value  []byte
}

// New creates a new disk cache in the given directory.
func New(dir string) (Cache, error) {
	var err error
	// Validate the directory.
	if len(dir) == 0 {
		return Cache{}, fmt.Errorf("directory path is empty")
	}
	// Create the directory if it doesn't exist.
	// MkdirAll creates a directory and any necessary parents and
	// is a no-op if the directory already exists.
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return Cache{}, fmt.Errorf("error creating cache directory: %w", err)
	}
	return Cache{dir: dir}, nil
}

// Delete removes the cache directory and all its contents.
func (c Cache) Delete() error {
	return os.RemoveAll(c.dir)
}

// Dir returns the directory path of the cache.
func (c Cache) Dir() string {
	return c.dir
}

// Filename returns the filename of a cache entry.
// TODO: Remove Filename from the public API?
func (c Cache) Filename(key string) string {
	return fmt.Sprintf("%x.json", sha256.Sum256([]byte(key)))
}

// Filepath returns the full path of a cache entry.
// TODO: Remove Filepath from the public API?
func (c Cache) Filepath(key string) string {
	return c.filepath(c.Filename(key))
}

// Set saves a cache entry with a key, value, and duration.
func (c Cache) Set(key string, value []byte, duration time.Duration) error {
	// Validate the key.
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	bytes, err := json.Marshal(Data{
		Key:    key,
		Value:  value,
		Expiry: time.Now().Add(duration),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(c.Filepath(key), bytes, 0644)
}

// Read reads a cache entry from disk and returns all its data.
// It does not check if the entry is expired.
func (c Cache) Read(key string) (Data, error) {
	return c.readFile(c.Filename(key))
}

// Has checks if a cache entry exists on disk.
func (c Cache) Has(key string) bool {
	_, err := os.Stat(c.Filepath(key))
	return err == nil
}

// Get gets a cache entry from disk and returns the value only.
// It returns an error if the entry is expired.
func (c Cache) Get(key string) ([]byte, error) {
	entry, err := c.Read(key)
	if err != nil {
		return nil, err
	}
	if time.Now().After(entry.Expiry) {
		return nil, fmt.Errorf("cache expired")
	}
	return entry.Value, nil
}

// Expiry returns the expiry time of a cache entry.
func (c Cache) Expiry(key string) time.Time {
	entry, err := c.Read(key)
	if err != nil {
		return time.Time{}
	}
	return entry.Expiry
}

// IsExpired returns true if a cache entry is expired.
func (c Cache) IsExpired(key string) bool {
	return time.Now().After(c.Expiry(key))
}

func (c Cache) list() ([]Data, error) {
	dirEntries, err := os.ReadDir(c.dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}
	var list []Data
	for _, dirEntry := range dirEntries {
		entry, err := c.readDirEntry(dirEntry)
		if err != nil {
			return nil, fmt.Errorf("error reading entry: %w", err)
		}
		list = append(list, entry)
	}
	return list, nil
}

// List returns a list of cache entry data.
// It accepts sorting options.
func (c Cache) List(options ...func([]Data)) ([]Data, error) {
	list, err := c.list()
	if err != nil {
		return nil, err
	}
	// Apply the sorting options.
	for _, option := range options {
		option(list)
	}
	return list, nil
}

// SortByExpiry is a sort function to sort cache entries by expiry time.
func SortByExpiry(entries []Data) {
	slices.SortFunc(entries, func(a, b Data) int {
		switch {
		case a.Expiry.Before(b.Expiry):
			return -1
		case a.Expiry.After(b.Expiry):
			return 1
		default:
			return 0
		}
	})
}

// SortByKey is a sort function to sort cache entries by key.
func SortByKey(entries []Data) {
	slices.SortFunc(entries, func(a, b Data) int {
		return strings.Compare(a.Key, b.Key)
	})
}

// SortByValue is a sort function to sort cache entries by value.
func SortByValue(entries []Data) {
	slices.SortFunc(entries, func(a, b Data) int {
		return strings.Compare(string(a.Value), string(b.Value))
	})
}

// Flush deletes all cache entries from disk.
func (c Cache) Flush() error {
	dirEntries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	var errs error
	for _, dirEntry := range dirEntries {
		err = c.removeDirEntry(dirEntry)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return errs
	}
	return nil
}

// Clean deletes expired cache entries from disk.
func (c Cache) Clean() error {
	var errs error
	list, err := c.list()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(list))
	for _, data := range list {
		wg.Add(1)
		go func(data Data) {
			defer wg.Done()
			if time.Now().Before(data.Expiry) {
				return
			}
			err := c.Remove(data.Key)
			if err != nil {
				errorsChan <- err
			}
		}(data)
	}
	wg.Wait()
	close(errorsChan)
	for err := range errorsChan {
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// Remove deletes a cache entry from disk.
func (c Cache) Remove(key string) error {
	return os.Remove(c.Filepath(key))
}

// readDirEntry reads an entry from disk.
// It differs from the Read method in that it takes a fs.DirEntry instead of a key.
// It's not part of the public API because the filename is not known outside the package.
func (c Cache) readDirEntry(dirEntry fs.DirEntry) (Data, error) {
	return c.readFile(dirEntry.Name())
}

// readFile reads a cache entry from disk.
// It takes a filename instead of a key.
func (c Cache) readFile(filename string) (Data, error) {
	bytes, err := os.ReadFile(c.filepath(filename))
	if err != nil {
		return Data{}, fmt.Errorf("error reading data: %w", err)
	}
	var entry Data
	err = json.Unmarshal(bytes, &entry)
	if err != nil {
		return Data{}, fmt.Errorf("error unmarshaling data: %w", err)
	}
	return entry, nil
}

// filepath returns the full path of a cache entry.
func (c Cache) filepath(filename string) string {
	return filepath.Join(c.dir, filename)
}

// removeFile deletes a cache entry from disk.
func (c Cache) removeFile(filename string) error {
	return os.Remove(c.filepath(filename))
}

// removeDirEntry deletes a cache entry from disk.
// It differs from the Remove method in that it takes a fs.DirEntry instead of a key.
func (c Cache) removeDirEntry(dirEntry fs.DirEntry) error {
	return c.removeFile(dirEntry.Name())
}
