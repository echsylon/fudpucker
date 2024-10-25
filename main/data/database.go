package data

import (
	"bytes"
	"errors"

	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/unit"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrConnection = errors.New("connection error")
	ErrIdFormat   = errors.New("invalid key")
)

type Database interface {
	Get(id entity.Id, attribute entity.Id) (map[entity.Id][]byte, error)
	Set(id entity.Id, data map[entity.Id][]byte) error
	Delete(id entity.Id) error
}

type database struct {
	options badger.Options
}

func NewDiskDatabase(path string) Database {
	return &database{options: buildOptions(path, false)}
}

func NewMemoryDatabase() Database {
	return &database{options: buildOptions("", true)}
}

// Get returns the requested attribute for the item with the
// given id. If something goes wrong, the raw, database
// implementation specific error is propagated.
func (d *database) Get(id entity.Id, attribute entity.Id) (map[entity.Id][]byte, error) {
	key, keyErr := buildDatabaseKey(id, attribute)
	if keyErr != nil {
		return nil, ErrIdFormat
	}

	db, dbErr := badger.Open(d.options)
	if dbErr != nil {
		return nil, dbErr
	}

	defer db.Close()
	data := make(map[entity.Id][]byte)
	cause := db.View(func(transaction *badger.Txn) error {
		if id != entity.ZeroId && attribute != entity.ZeroId {
			return copySingleAttribute(transaction, key, attribute, data)
		} else {
			return copyAllAttributes(transaction, key, data)
		}
	})

	return data, cause
}

// Set writes (puts or updates) all given attributes data to
// the database. If something goes wrong, the raw, database
// implementation specific error is propagated.
func (d *database) Set(id entity.Id, data map[entity.Id][]byte) error {
	db, dbErr := badger.Open(d.options)
	if dbErr != nil {
		return dbErr
	}

	defer db.Close()
	return db.Update(func(transaction *badger.Txn) error {
		// Ensure we have the entity id stored as well under the key <zero_id><entity_id>.
		// This is important when we want to list all stored entities. To do that we will
		// search for all entries with key prefix <zero_id>.
		if idKey, keyErr := buildDatabaseKey(entity.ZeroId, id); keyErr != nil {
			return keyErr
		} else if setErr := transaction.Set(idKey, []byte{}); setErr != nil {
			return setErr
		}

		// Now, store all entity attributes under respective <entity_id><attribute> key.
		for attr, value := range data {
			if key, keyErr := buildDatabaseKey(id, attr); keyErr != nil {
				return ErrIdFormat
			} else if setErr := transaction.Set(key, value); setErr != nil {
				return setErr
			}
		}

		return nil
	})
}

// Delete removes all attributes with a matching prefix from
// the database. If something goes wrong, the raw, database
// implementation specific error is returned.
func (d *database) Delete(id entity.Id) error {
	prefix, keyErr := buildDatabaseKey(id, entity.ZeroId)
	if keyErr != nil {
		return ErrIdFormat
	}

	key, keyErr := buildDatabaseKey(entity.ZeroId, id)
	if keyErr != nil {
		return ErrIdFormat
	}

	db, dbErr := badger.Open(d.options)
	if dbErr != nil {
		return dbErr
	}

	defer db.Close()
	return db.DropPrefix(key, prefix)
}

// Private helper functions
func buildOptions(path string, inMemory bool) badger.Options {
	return badger.DefaultOptions(path).
		WithIndexCacheSize(10 * unit.MiB).
		WithMetricsEnabled(false).
		WithNumVersionsToKeep(1).
		WithInMemory(inMemory).
		WithLogger(nil)
}

func buildDatabaseKey(id entity.Id, attribute entity.Id) ([]byte, error) {
	if attribute == entity.ZeroId {
		return id.Bytes(), nil
	} else {
		writer := bytes.NewBuffer([]byte{})
		writer.Write(id.Bytes())
		writer.Write(attribute.Bytes())
		return writer.Bytes(), nil
	}
}

func copySingleAttribute(transaction *badger.Txn, key []byte, attribute entity.Id, result map[entity.Id][]byte) error {
	if item, err := transaction.Get(key); err != nil {
		return err
	} else if result[attribute], err = item.ValueCopy(nil); err != nil {
		return err
	} else {
		return nil
	}
}

func copyAllAttributes(transaction *badger.Txn, prefix []byte, result map[entity.Id][]byte) error {
	options := badger.DefaultIteratorOptions
	options.PrefetchValues = false // find fast but read a bit slower
	options.AllVersions = false    // only get the last written value
	iterator := transaction.NewIterator(options)
	defer iterator.Close()

	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		attrIndex := len(prefix)
		attrBytes := item.Key()[attrIndex:]
		if attr, err := entity.NewBytesId(attrBytes); err != nil {
			continue
		} else if bytes, err := item.ValueCopy(nil); err != nil {
			continue
		} else {
			result[attr] = bytes
		}
	}

	return nil
}
