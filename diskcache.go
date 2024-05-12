package diskcache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

type Cache struct {
	dir string
}

type data struct {
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
	bytes, err := json.Marshal(data{
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

func (c Cache) Load(key string) (data, error) {
	bytes, err := os.ReadFile(c.Path(key))
	if err != nil {
		return data{}, err
	}
	var d data
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return data{}, err
	}
	return d, nil
}

func (c Cache) Remove(key string) error {
	return os.Remove(c.Path(key))
}

func (c Cache) Clear() error {
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

func (c Cache) List() ([]string, error) {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, file := range files {
		bytes, err := os.ReadFile(path.Join(c.dir, file.Name()))
		if err != nil {
			return nil, err
		}
		var d data
		err = json.Unmarshal(bytes, &d)
		if err != nil {
			return nil, err
		}
		keys = append(keys, d.Key)
	}
	return keys, nil
}
