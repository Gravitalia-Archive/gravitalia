package database

import (
	"os"

	"github.com/bradfitz/gomemcache/memcache"
)

var Mem = memcache.New(os.Getenv("MEM_URL"))

// Set permits to set a temporary value, on the cache
// via Memcached
func Set(key string, value string, time int32) {
	Mem.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: time,
	})
}
