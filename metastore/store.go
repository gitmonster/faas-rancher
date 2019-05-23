package metastore

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	bolt "go.etcd.io/bbolt"
)

const (
	storeRoot = "/metastore"
)

var (
	database            *bolt.DB
	bucketNameFunctions = []byte("functions")
)

var (
	ErrEntityNotFound         = errors.New("entity not found")
	ErrDatabaseNotInitialized = errors.New("database not initialized")
	ErrServiceNameUndefined   = errors.New("serviceNameUndefined")
)

//Open opens a database
func Open() error {
	db, err := bolt.Open(fmt.Sprintf("%s/store.db", storeRoot), 0600, nil)
	if err != nil {
		return errors.Annotate(err, "Open")
	}

	database = db
	return database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketNameFunctions)
		if err != nil {
			return errors.Annotate(err, "CreateBucketIfNotExists")
		}
		return nil
	})
}

// Close closes the database
func Close() (err error) {
	if database != nil {
		err = database.Close()
		database = nil
	}
	return err
}

// Update updates metadata related to a service
func Update(meta *FunctionMeta) error {
	if database == nil {
		return ErrDatabaseNotInitialized
	}

	if meta.Service == "" {
		return ErrServiceNameUndefined
	}

	return database.Update(func(tx *bolt.Tx) error {
		buf, err := json.Marshal(meta)
		if err != nil {
			return errors.Annotate(err, "Marshal")
		}

		b := tx.Bucket(bucketNameFunctions)
		return b.Put([]byte(meta.Service), buf)
	})
}

// Read reads metadata related to a service
func Read(meta *FunctionMeta) error {
	if database == nil {
		return ErrDatabaseNotInitialized
	}

	if meta.Service == "" {
		return ErrServiceNameUndefined
	}

	return database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameFunctions)
		buf := b.Get([]byte(meta.Service))
		if buf == nil {
			return ErrEntityNotFound
		}

		if err := json.Unmarshal(buf, meta); err != nil {
			return errors.Annotate(err, "Unmarshal")
		}

		return nil
	})
}
