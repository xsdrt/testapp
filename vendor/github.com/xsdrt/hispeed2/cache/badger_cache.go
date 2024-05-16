package cache

import (
	"time"

	"github.com/dgraph-io/badger/v3"
)

// From BadgerDB Repository:"BadgerDB is an embeddable, persistent and fast key-value (KV) database"
// "written in pure Go. It is the underlying database for Dgraph, a fast, distributed graph database"
type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

// A simple check if a given key exist in BadgerDB
func (b *BadgerCache) Has(str string) (bool, error) {
	_, err := b.Get(str)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Get somthing out of the cache...
func (b *BadgerCache) Get(str string) (interface{}, error) {
	var fromCache []byte

	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			fromCache = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}

	item := decoded[str]

	return item, nil
}

// Put(set) something in the cache...
func (b *BadgerCache) Set(str string, value interface{}, expires ...int) error {
	entry := Entry{}

	entry[str] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded).WithTTL(time.Second * time.Duration(expires[0]))
			err = txn.SetEntry(e)
			return err

		})
	} else {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded)
			err = txn.SetEntry(e)
			return err

		})

	}

	return nil
}

// Forget something in the cache...
func (b *BadgerCache) Forget(str string) error {
	err := b.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(str))
		return err
	})

	return err
}

// Empty the cache using a matching pattern...
func (b *BadgerCache) EmptyByMatch(str string) error {
	return b.emptyByMatch(str)
}

// Empty the entire cache...
func (b *BadgerCache) Empty() error {
	return b.emptyByMatch("")
}

func (b *BadgerCache) emptyByMatch(str string) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := b.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000

	err := b.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		iter := txn.NewIterator(opts)
		defer iter.Close() // make sure and close the Iterator when done...

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for iter.Seek([]byte(str)); iter.ValidForPrefix([]byte(str)); iter.Next() {
			key := iter.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize { // If keys are at 100,000 then delete them, more takes time...
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
			}

		}
		if keysCollected > 0 { // check if some left after deleting 100,000 or if less and then delete them
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
