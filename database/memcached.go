package database

import "github.com/bradfitz/gomemcache/memcache"

var Mem = memcache.New("127.0.0.1:11211")

func Set(key string, value string) {
	Mem.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: 300,
	})
}
