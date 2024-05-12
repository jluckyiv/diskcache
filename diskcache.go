package diskcache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

type Cache struct {
	dir string
}

type Entry struct {
	Expiry time.Time
	Key    string
	Value  []byte
}

func New(dir string) (Cache, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return Cache{}, err
	}
	return Cache{
		dir: dir,
	}, nil
}

func (c Cache) Dir() string {
	return c.dir
}

func (c Cache) Filename(key string) string {
	return fmt.Sprintf("%x.json", sha256.Sum256([]byte(key)))
}

func (c Cache) Path(key string) string {
	return path.Join(c.dir, c.Filename(key))
}

func (c Cache) Put(key string, value []byte, duration time.Duration) error {
	bytes, err := json.Marshal(Entry{
		Key:    key,
		Value:  value,
		Expiry: time.Now().Add(duration),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path(key), bytes, 0644)
}

func (c Cache) Get(key string) ([]byte, error) {
	d, err := c.Load(key)
	if err != nil {
		return nil, err
	}
	if time.Now().After(d.Expiry) {
		return nil, fmt.Errorf("cache expired")
	}
	return d.Value, nil
}

func (c Cache) Expiry(key string) time.Time {
	d, err := c.Load(key)
	if err != nil {
		return time.Time{}
	}
	return d.Expiry
}

func (c Cache) IsExpired(key string) bool {
	return time.Now().After(c.Expiry(key))
}

func (c Cache) Load(key string) (Entry, error) {
	bytes, err := os.ReadFile(c.Path(key))
	if err != nil {
		return Entry{}, err
	}
	var d Entry
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return Entry{}, err
	}
	return d, nil
}

func (c Cache) Remove(key string) error {
	return os.Remove(c.Path(key))
}

func (c Cache) Flush() error {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.Remove(path.Join(c.dir, file.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Cache) Clean() error {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		filename := path.Join(c.dir, file.Name())
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		var d Entry
		err = json.Unmarshal(bytes, &d)
		if err != nil {
			return err
		}
		if time.Now().Before(d.Expiry) {
			continue
		}
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func SortByKey(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		return strings.Compare(a.Key, b.Key)
	})
}

func SortByValue(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		return strings.Compare(string(a.Value), string(b.Value))
	})
}

func SortByExpiry(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
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

func (c Cache) List(options ...func([]Entry)) ([]Entry, error) {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, file := range files {
		bytes, err := os.ReadFile(path.Join(c.dir, file.Name()))
		if err != nil {
			return nil, err
		}
		var d Entry
		err = json.Unmarshal(bytes, &d)
		if err != nil {
			return nil, err
		}
		entries = append(entries, d)
	}

	// Apply the sorting options.
	for _, option := range options {
		option(entries)
	}
	return entries, nil
}
