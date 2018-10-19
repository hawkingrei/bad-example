package memcache_test

import (
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/memcache"
)

func ExampleConn_set() {
	var (
		err    error
		value  []byte
		conn   memcache.Conn
		expire int32 = 100
		p            = struct {
			Name string
			Age  int64
		}{"golang", 10}
	)
	cnop := memcache.DialConnectTimeout(time.Duration(time.Second))
	rdop := memcache.DialReadTimeout(time.Duration(time.Second))
	wrop := memcache.DialWriteTimeout(time.Duration(time.Second))
	if value, err = json.Marshal(p); err != nil {
		fmt.Println(err)
		return
	}
	if conn, err = memcache.Dial("tcp", "172.16.33.54:11211", cnop, rdop, wrop); err != nil {
		fmt.Println(err)
		return
	}
	// FlagRAW test
	itemRaw := &memcache.Item{
		Key:        "test_raw",
		Value:      value,
		Expiration: expire,
	}
	if err = conn.Set(itemRaw); err != nil {
		fmt.Println(err)
		return
	}
	// FlagGzip
	itemGZip := &memcache.Item{
		Key:        "test_gzip",
		Value:      value,
		Flags:      memcache.FlagGzip,
		Expiration: expire,
	}
	if err = conn.Set(itemGZip); err != nil {
		fmt.Println(err)
		return
	}
	// FlagGOB
	itemGOB := &memcache.Item{
		Key:        "test_gob",
		Object:     p,
		Flags:      memcache.FlagGOB,
		Expiration: expire,
	}
	if err = conn.Set(itemGOB); err != nil {
		fmt.Println(err)
		return
	}
	// FlagJSON
	itemJSON := &memcache.Item{
		Key:        "test_json",
		Object:     p,
		Flags:      memcache.FlagJSON,
		Expiration: expire,
	}
	if err = conn.Set(itemJSON); err != nil {
		fmt.Println(err)
		return
	}
	// FlagJSON | FlagGzip
	itemJSONGzip := &memcache.Item{
		Key:        "test_jsonGzip",
		Object:     p,
		Flags:      memcache.FlagJSON | memcache.FlagGzip,
		Expiration: expire,
	}
	if err = conn.Set(itemJSONGzip); err != nil {
		fmt.Println(err)
		return
	}
	// Output:
}

func ExampleConn_get() {
	var (
		err  error
		item *memcache.Item
		conn memcache.Conn
		p    struct {
			Name string
			Age  int64
		}
	)
	cnop := memcache.DialConnectTimeout(time.Duration(time.Second))
	rdop := memcache.DialReadTimeout(time.Duration(time.Second))
	wrop := memcache.DialWriteTimeout(time.Duration(time.Second))
	if conn, err = memcache.Dial("tcp", "172.16.33.54:11211", cnop, rdop, wrop); err != nil {
		fmt.Println(err)
		return
	}
	if item, err = conn.Get("test_raw"); err != nil {
		fmt.Println(err)
	} else {
		if err = conn.Scan(item, &p); err != nil {
			fmt.Printf("FlagRAW conn.Scan error(%v)\n", err)
			return
		}
	}
	// FlagGZip
	if item, err = conn.Get("test_gzip"); err != nil {
		fmt.Println(err)
	} else {
		if err = conn.Scan(item, &p); err != nil {
			fmt.Printf("FlagGZip conn.Scan error(%v)\n", err)
			return
		}
	}
	// FlagGOB
	if item, err = conn.Get("test_gob"); err != nil {
		fmt.Println(err)
	} else {
		if err = conn.Scan(item, &p); err != nil {
			fmt.Printf("FlagGOB conn.Scan error(%v)\n", err)
			return
		}
	}
	// FlagJSON
	if item, err = conn.Get("test_json"); err != nil {
		fmt.Println(err)
	} else {
		if err = conn.Scan(item, &p); err != nil {
			fmt.Printf("FlagJSON conn.Scan error(%v)\n", err)
			return
		}
	}
	// Output:
}

func ExampleConn_getMulti() {
	var (
		err  error
		conn memcache.Conn
		res  map[string]*memcache.Item
		keys = []string{"test_raw", "test_gzip"}
		p    struct {
			Name string
			Age  int64
		}
	)
	cnop := memcache.DialConnectTimeout(time.Duration(time.Second))
	rdop := memcache.DialReadTimeout(time.Duration(time.Second))
	wrop := memcache.DialWriteTimeout(time.Duration(time.Second))
	if conn, err = memcache.Dial("tcp", "172.16.33.54:11211", cnop, rdop, wrop); err != nil {
		fmt.Println(err)
		return
	}
	if res, err = conn.GetMulti(keys); err != nil {
		fmt.Printf("conn.GetMulti(%v) error(%v)", keys, err)
		return
	}
	for _, v := range res {
		if err = conn.Scan(v, &p); err != nil {
			fmt.Printf("conn.Scan error(%v)\n", err)
			return
		}
		fmt.Println(p)
	}
	// Output:
	//{golang 10}
	//{golang 10}
}
