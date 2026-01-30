package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// BucketName is the name of the BoltDB bucket used to store secrets
const BucketName = "secrets"

// ErrBucketNotFound is the error message when the secrets bucket is not found
const ErrBucketNotFound = "bucket not found"

// Secret represents a stored encrypted secret with timestamp
type Secret struct {
	Timestamp int64  `json:"timestamp"`
	Encrypted string `json:"encrypted"` // base64 encoded
}

// InitBucket creates the secrets bucket if it doesn't exist
func InitBucket(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		return err
	})
}

// Store saves an encrypted secret with timestamp
func Store(db *bolt.DB, key string, encrypted []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return errors.New(ErrBucketNotFound)
		}

		encoded := base64.StdEncoding.EncodeToString(encrypted)
		secret := Secret{
			Timestamp: time.Now().UnixMilli(),
			Encrypted: encoded,
		}

		data, err := json.Marshal(secret)
		if err != nil {
			return fmt.Errorf("failed to marshal secret: %w", err)
		}

		return b.Put([]byte(key), data)
	})
}

// Retrieve gets a secret by key (does NOT delete it)
func Retrieve(db *bolt.DB, key string) (*Secret, error) {
	var secret Secret
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return errors.New(ErrBucketNotFound)
		}

		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &secret)
	})

	if err != nil {
		return nil, err
	}

	// If Timestamp is 0, it means the secret was empty/not found (because json unmarshal didn't run or data was empty)
	// But Get() returns nil if not found, and we return nil error there.
	// So if secret.Encrypted is empty and Timestamp is 0, we can assume it wasn't found.
	if secret.Timestamp == 0 && secret.Encrypted == "" {
		return nil, nil
	}

	return &secret, nil
}

// Delete removes a secret by key (burn operation)
func Delete(db *bolt.DB, key string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return errors.New(ErrBucketNotFound)
		}
		return b.Delete([]byte(key))
	})
}

// DeleteExpired removes all secrets older than ttl (in days)
func DeleteExpired(db *bolt.DB, ttlDays int) (int, error) {
	count := 0
	cutoff := time.Now().Add(time.Duration(-ttlDays) * 24 * time.Hour).UnixMilli()

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return errors.New(ErrBucketNotFound)
		}

		var keysToDelete [][]byte

		err := b.ForEach(func(k, v []byte) error {
			var secret Secret
			if err := json.Unmarshal(v, &secret); err != nil {
				return nil // Skip invalid JSON entries
			}

			if secret.Timestamp < cutoff {
				// We must copy the key because k is only valid for the current iteration
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
			}
			return nil
		})
		if err != nil {
			return err
		}

		for _, k := range keysToDelete {
			if b.Delete(k) == nil {
				count++
			}
		}

		return nil
	})

	return count, err
}
