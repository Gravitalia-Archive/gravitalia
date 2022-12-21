package database

import "github.com/bradfitz/gomemcache/memcache"

var Mem = memcache.New("127.0.0.1:11211")
