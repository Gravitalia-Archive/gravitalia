package database

import "github.com/bradfitz/gomemcache/memcache"

var Mem = memcache.New("127.0.0.1:11211")

// Set permits to set a temporary value (5 minutes), on the cache
// via Memcached
func Set(key string, value string) {
	Mem.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: 300,
	})
}
