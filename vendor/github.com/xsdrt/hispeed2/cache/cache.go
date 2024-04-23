package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type Cache interface { // List all the functions must implement to satisfy this interface (go's interface system is so simple its great!!)
	Has(string) (bool, error)              // Does the cache have something (string/the key)
	Get(string) (interface{}, error)       // Use the string/key and a interface since a cache can store/return anything, also a potenial error..
	Set(string, interface{}, ...int) error // Set what we want to store (string/key;...int is where we set the expirery for the cache I.E. 60 secs or 25,000secs etc...
	Forget(string) error                   // Take it out of (Forget) cache by key(whatever you named the value/key)
	EmptyByMatch(string) error             // Everything in the cache by a pattern I.E. everything that starts with letter a , or whatever...
	Empty() error                          // Just empty everything no paremeters, just everything...
}

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string // use prefix(unique);case two or more app(s) use redis w/same key i.e. tempid then Forget func and all cached keys gone now...
}

type Entry map[string]interface{} // a map of serialized item(s) to pull out and deserialize...

// A simple check if a given key exist in redis cache for refernce on how works TODO: Need to implement the hash...
func (c *RedisCache) Has(str string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str) // prepend prefix to the key(str), so i.e. <prefix>:<whatever the user gave>
	conn := c.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return ok, nil
}

// Serialize and deserialize for the cache
func encode(item Entry) ([]byte, error) { // No reciever, we want the ability to call other cache types from anywhere in the is pkg...
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decode(str string) (Entry, error) {
	item := Entry{}
	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err := d.Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// Get somthing out of the cache...
func (c *RedisCache) Get(str string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str) // prepend prefix to the key(str), so i.e. <prefix>:<whatever the user gave>
	conn := c.Conn.Get()
	defer conn.Close()

	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(cacheEntry))
	if err != nil {
		return nil, err
	}

	item := decoded[key]

	return item, nil // Just returning anything and  nil for now
}

// Put(set) something in the cache...
func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str) // prepend prefix to the key(str), so i.e. <prefix>:<whatever the user gave>
	conn := c.Conn.Get()
	defer conn.Close()

	entry := Entry{}
	entry[key] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		_, err := conn.Do("SETEX", key, expires[0], string(encoded))
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Do("SETEX", key, string(encoded))
		if err != nil {
			return err
		}

	}

	return nil
}

// Forget something in the cache...
func (c *RedisCache) Forget(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str) // prepend prefix to the key(str), so i.e. <prefix>:<whatever the user gave>
	conn := c.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

// Empty the cache using a matching pattern...
func (c *RedisCache) EmptyByMatch(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str) // prepend prefix to the key(str), so i.e. <prefix>:<whatever the user gave>
	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		err := c.Forget(x)
		if err != nil {
			return err
		}
	}

	return nil
}

// Empty the entire cache...
func (c *RedisCache) Empty() error {
	key := fmt.Sprintf("%s:", c.Prefix) // Want get rid of stuff in the cache htat has the prefix: (notice colon) appened to them...
	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		err = c.Forget(x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := c.Conn.Get()
	defer conn.Close()

	iter := 0
	keys := []string{} // Populate this slice of strings with all things we need to get rid of...
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return keys, err
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
