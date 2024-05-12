package diskcache_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jluckyiv/diskcache"
)

func TestDiskCache(t *testing.T) {
	tempdir := t.TempDir()
	cacheFolder := "testcache"
	cacheDir := path.Join(tempdir, cacheFolder)
	cache, err := diskcache.New(cacheDir)
	if err != nil {
		t.Fatalf("Error creating cache: %v", err)
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatalf("Cache dir %s does not exist", cache.Dir())
	}
	if cache.Dir() != cacheDir {
		t.Fatalf("Want cache dir to be %s, got %s", cacheDir, cache.Dir())
	}

	t.Run("TestFilename", func(t *testing.T) {
		key := "testkey"
		got := cache.Filename(key)
		want := fmt.Sprintf("%x.json", sha256.Sum256([]byte(key)))
		if got != want {
			t.Fatalf("Want filename to be %s, got %s", want, got)
		}
	})

	t.Run("TestPath", func(t *testing.T) {
		key := "testkey"
		got := cache.Path(key)
		filename := fmt.Sprintf("%x.json", sha256.Sum256([]byte(key)))
		want := path.Join(cacheDir, filename)
		if got != want {
			t.Fatalf("Want cache path to be %s, got %s", want, got)
		}
	})

	t.Run("TestData", func(t *testing.T) {
		key := "testkey"
		want := []byte("testvalue")
		err := cache.Put(key, want, 1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		got, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Error getting cache: %v", err)
		}
		if string(got) != string(want) {
			t.Fatalf("Expected cache value to be %s, got %s", string(want), string(got))
		}
		data, err := cache.Load(key)
		if err != nil {
			t.Fatalf("Error loading cache: %v", err)
		}
		if string(data.Value) != string(want) {
			t.Fatalf("Expected cache value to be %s, got %s", want, data.Value)
		}
		if data.Key != key {
			t.Fatalf("Expected cache key to be %s, got %s", key, data.Key)
		}
		if data.Expiry.IsZero() {
			t.Fatalf("Expected cache expiry to be non-zero")
		}
		if data.Expiry.Before(time.Now()) {
			t.Fatalf("Expected cache expiry to be in the future")
		}
		if data.Expiry.After(time.Now().Add(1 * time.Minute)) {
			t.Fatalf("Expected cache expiry to be within 1 minute")
		}
	})

	t.Run("TestUnexpiredCache", func(t *testing.T) {
		key := "unexpired"
		err := cache.Put(key, []byte(""), 1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		_, err = cache.Get(key)
		if err != nil {
			t.Fatalf("Error getting cache: %v", err)
		}
		expiry := cache.Expiry(key)
		if expiry.IsZero() {
			t.Fatalf("Expected cache expiry to be non-zero")
		}
		if expiry.Before(time.Now()) {
			t.Fatalf("Expected cache expiry to be in the future")
		}
		if expiry.After(time.Now().Add(1 * time.Minute)) {
			t.Fatalf("Expected cache expiry to be within 1 minute")
		}
		isExpired := cache.IsExpired(key)
		if isExpired {
			t.Fatalf("Expected cache to not be expired")
		}
	})

	t.Run("TestExpiredCache", func(t *testing.T) {
		key := "expired"
		err := cache.Put(key, []byte(""), -1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		_, err = cache.Get(key)
		if err == nil {
			t.Fatalf("Expected error getting cache")
		}
		if err.Error() != "cache expired" {
			t.Fatalf("Expected error message to be 'cache expired', got %s", err.Error())
		}
		isExpired := cache.IsExpired(key)
		if !isExpired {
			t.Fatalf("Expected cache to be expired")
		}
	})

	t.Run("TestUpdate", func(t *testing.T) {
		key := "testkey"
		oldvalue := []byte("oldvalue")
		err := cache.Put(key, oldvalue, 1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		got, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Error getting cache: %v", err)
		}
		if string(got) != string(oldvalue) {
			t.Fatalf("Expected cache value to be %s, got %s", string(oldvalue), string(got))
		}
		newvalue := []byte("newvalue")
		err = cache.Put(key, newvalue, 1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		got, err = cache.Get(key)
		if err != nil {
			t.Fatalf("Error getting cache: %v", err)
		}
		if string(got) != string(newvalue) {
			t.Fatalf("Expected cache value to be %s, got %s", string(newvalue), string(got))
		}
	})

	t.Run("TestRemove", func(t *testing.T) {
		key := "delete"
		err := cache.Put(key, []byte("value"), 1*time.Minute)
		if err != nil {
			t.Fatalf("Error saving cache: %v", err)
		}
		_, err = cache.Get(key)
		if err != nil {
			t.Fatalf("Error getting cache: %v", err)
		}
		err = cache.Remove(key)
		if err != nil {
			t.Fatalf("Error deleting cache: %v", err)
		}
		_, err = cache.Get(key)
		if err == nil {
			t.Fatalf("Expected error getting cache")
		}
	})

	t.Run("TestList", func(t *testing.T) {
		// Clear the cache.
		err := cache.Clear()
		if err != nil {
			t.Fatalf("Error clearing cache: %v", err)
		}

		empty, err := cache.List()
		if err != nil {
			t.Fatalf("Error listing cache: %v", err)
		}
		if len(empty) != 0 {
			t.Fatalf("Expected 0 keys, got %d", len(empty))
		}

		// Save some test data.
		testData := []struct {
			key   string
			value string
		}{
			{"key1", "value1"},
			{"key2", "value2"},
			{"key3", "value3"},
		}

		for _, td := range testData {
			err := cache.Put(td.key, []byte(td.value), 1*time.Minute)
			if err != nil {
				t.Fatalf("Error saving cache: %v", err)
			}
		}

		// List the keys.
		keys, err := cache.List()
		if err != nil {
			t.Fatalf("Error listing cache: %v", err)
		}
		if len(keys) != 3 {
			t.Fatalf("Expected 3 keys, got %d", len(keys))
		}

		// Check that all the test data keys are in the list.
		for _, td := range testData {
			found := false
			for _, key := range keys {
				if key == td.key {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Expected key %s to be in list", td.key)
			}
		}
	})
}
