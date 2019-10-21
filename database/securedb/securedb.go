// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package securedb

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
)

var (
	// Bucket for all db metadata
	EncyptedMetaData = []byte("EncyptedDBMetaData")

	challenge = []byte("Challenge")

	lockedError = fmt.Errorf("Wallet is locked")
	TimeCheck   = time.Time{} // Used to Check if time has been initiated
)

// EncryptedDB is a database with symmetric encryption to encrypt all writes, and decrypt all reads
type EncryptedDB struct {
	// Stores all encrypted data
	db interfaces.IDatabase

	// Metadata includes salt
	metadata *SecureDBMetaData

	// encryptionkey is a hash of the password and salt
	encryptionkey []byte

	// Allow the wallet to be locked, by gating access based
	// on time.
	UnlockedUntil time.Time
}

// NewEncryptedDB takes the filename, dbtype, and password.
//		Dbtype :
//			Map
//			Bolt
//			LevelDB
func NewEncryptedDB(filename, dbtype, password string) (*EncryptedDB, error) {
	e := new(EncryptedDB)
	e.Init(filename, dbtype)

	err := e.initSecureDB(password)
	if err != nil {
		e.Close()
		return nil, err
	}

	return e, nil
}

// InitSecureDB will init the Salt and metadata
func (db *EncryptedDB) initSecureDB(password string) error {
	m := new(SecureDBMetaData)
	v, err := db.db.Get(EncyptedMetaData, EncyptedMetaData, m)
	if err != nil {
		return err
	}

	if v == nil {
		// need to init new metadata
		db.initNewMetaData()
	} else {
		db.metadata = m
	}

	key, err := GetKey(password, db.metadata.Salt.Bytes)
	if err != nil {
		return err
	}

	db.encryptionkey = key

	if len(db.metadata.Challenge.Bytes) == 0 {
		// Create challenge
		cipherText, err := Encrypt(challenge, db.encryptionkey)
		if err != nil {
			return err
		}

		var c primitives.ByteSlice
		c.Bytes = cipherText
		db.metadata.Challenge = c
		err = db.db.Put(EncyptedMetaData, EncyptedMetaData, db.metadata)
		if err != nil {
			return err
		}

	} else {
		// Do challenge
		plainText, err := Decrypt(db.metadata.Challenge.Bytes, db.encryptionkey)
		if err != nil {
			return fmt.Errorf("password supplied is incorrect, and cannot decrypt the existing database")
		}

		if subtle.ConstantTimeCompare(plainText, challenge) == 0 {
			return fmt.Errorf("password supplied is incorrect, and cannot decrypt the existing database")
		}
	}

	return nil
}

func (db *EncryptedDB) initNewMetaData() {
	db.metadata = NewSecureDBMetaData()
	salt := make([]byte, 30)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}

	db.metadata.Salt.Bytes = salt
}

/***************************************
 *       Methods
 ***************************************/

// Lock locks the database such that further access calls will fail until it is
// unlocked
func (db *EncryptedDB) Lock() {
	db.UnlockedUntil = time.Now().Add(-1 * time.Minute)
}

func (db *EncryptedDB) UnlockFor(password string, duration time.Duration) error {
	key, err := GetKey(password, db.metadata.Salt.Bytes)
	if err != nil {
		return err
	}
	if bytes.Compare(key, db.encryptionkey) != 0 {
		return fmt.Errorf("incorrect password")
	}

	db.UnlockedUntil = time.Now().Add(duration)
	return nil
}

func (db *EncryptedDB) isLocked() bool {
	// Not being used, means always unlocked
	if db.UnlockedUntil == TimeCheck {
		return false
	}

	// All times after the unlock time set, are invalid
	if time.Now().Before(db.UnlockedUntil) {
		return false
	}
	return true
}

func (db *EncryptedDB) ListAllBuckets() ([][]byte, error) {
	if db.isLocked() {
		return nil, lockedError
	}
	return db.db.ListAllBuckets()
}

// We don't care if delete works or not.  If the key isn't there, that's ok
func (db *EncryptedDB) Delete(bucket []byte, key []byte) error {
	if db.isLocked() {
		return lockedError
	}
	return db.db.Delete(bucket, key)
}

// Can't trim a real database
func (db *EncryptedDB) Trim() {
}

func (db *EncryptedDB) Close() error {
	return db.db.Close()
}

func (db *EncryptedDB) Get(bucket []byte, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	if db.isLocked() {
		return nil, lockedError
	}

	e := NewEncryptedMarshaler(db.encryptionkey, destination)
	tmp, err := db.db.Get(bucket, key, e)
	if err != nil {
		return nil, err
	}

	if tmp == nil {
		return nil, nil
	}

	return e.Original, nil
}

func (db *EncryptedDB) Put(bucket []byte, key []byte, data interfaces.BinaryMarshallable) error {
	if db.isLocked() {
		return lockedError
	}

	e := NewEncryptedMarshaler(db.encryptionkey, data)
	return db.db.Put(bucket, key, e)
}

func (db *EncryptedDB) PutInBatch(records []interfaces.Record) error {
	if db.isLocked() {
		return lockedError
	}

	cipherRecords := make([]interfaces.Record, len(records))
	for i, r := range records {
		cipherRecords[i].Bucket = r.Bucket
		cipherRecords[i].Key = r.Key

		e := NewEncryptedMarshaler(db.encryptionkey, r.Data)
		cipherRecords[i].Data = e
	}

	return db.db.PutInBatch(cipherRecords)
}

func (db *EncryptedDB) Clear(bucket []byte) error {
	if db.isLocked() {
		return lockedError
	}

	return db.db.Clear(bucket)
}

func (db *EncryptedDB) ListAllKeys(bucket []byte) (keys [][]byte, err error) {
	if db.isLocked() {
		return nil, lockedError
	}

	keys, err = db.db.ListAllKeys(bucket)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (db *EncryptedDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	if db.isLocked() {
		return nil, nil, lockedError
	}

	s := NewEncryptedMarshaler(db.encryptionkey, sample.(interfaces.BinaryMarshallable))

	cipheredAll, keys, err := db.db.GetAll(bucket, s)
	if err != nil {
		return nil, nil, err
	}

	originalSamples := make([]interfaces.BinaryMarshallableAndCopyable, len(cipheredAll))
	for i, e := range cipheredAll {
		originalSamples[i] = e.(*EncryptedMarshaler).Original.(interfaces.BinaryMarshallableAndCopyable)
	}

	return originalSamples, keys, err
}

func (db *EncryptedDB) Init(filename string, dbtype string) {
	var err error
	switch dbtype {
	case "Map":
		db.db = new(mapdb.MapDB)
	case "LDB":
		db.db, err = leveldb.NewLevelDB(filename, true)
		if err != nil {
			panic(err)
		}
	case "Bolt":
		db.db = boltdb.NewBoltDB(nil, filename)
	default:
		panic(fmt.Sprintf("%s is not a valid option. Expect 'Map', 'LDB', or 'Bolt'", dbtype))
	}
}

func (db *EncryptedDB) DoesKeyExist(bucket, key []byte) (bool, error) {
	if db.isLocked() {
		return false, fmt.Errorf("Wallet is locked")
	}

	return db.db.DoesKeyExist(bucket, key)
}
