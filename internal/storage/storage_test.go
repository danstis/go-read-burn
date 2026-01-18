package storage

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func setupTestDB(t *testing.T) *bolt.DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	// Use default BoltDB options which include mmap flags that might help with alignment
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestInitBucket(t *testing.T) {
	db := setupTestDB(t)

	err := InitBucket(db)
	if err != nil {
		t.Fatalf("InitBucket failed: %v", err)
	}

	// Verify bucket exists
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("secrets"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return nil
	})
	if err != nil {
		t.Errorf("Bucket 'secrets' was not created")
	}
}

func TestStoreAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	key := "testkey"
	data := []byte("secret data")
	
	// Test Store
	if err := Store(db, key, data); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Test Retrieve
	secret, err := Retrieve(db, key)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if secret == nil {
		t.Fatal("Retrieve returned nil secret")
	}

	// Verify data
	decoded, err := base64.StdEncoding.DecodeString(secret.Encrypted)
	if err != nil {
		t.Fatalf("Failed to decode stored secret: %v", err)
	}
	if string(decoded) != string(data) {
		t.Errorf("Got %s, want %s", string(decoded), string(data))
	}
}

func TestRetrieveNonExistent(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	secret, err := Retrieve(db, "nonexistent")
	if err != nil {
		t.Errorf("Retrieve returned error for missing key: %v", err)
	}
	if secret != nil {
		t.Error("Retrieve returned non-nil secret for missing key")
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	key := "todelete"
	data := []byte("data")
	Store(db, key, data)

	if err := Delete(db, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	secret, err := Retrieve(db, key)
	if err != nil {
		t.Errorf("Retrieve after delete failed: %v", err)
	}
	if secret != nil {
		t.Error("Secret still exists after delete")
	}
}

func TestDeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	// 1. Store a fresh secret
	freshKey := "fresh"
	Store(db, freshKey, []byte("data"))

	// 2. Manually store an expired secret (2 days old)
	expiredKey := "expired"
	expiredSecret := Secret{
		Timestamp: time.Now().Add(-48 * time.Hour).UnixMilli(),
		Encrypted: "expireddata",
	}
	expiredData, _ := json.Marshal(expiredSecret)
	
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("secrets"))
		return b.Put([]byte(expiredKey), expiredData)
	})
	if err != nil {
		t.Fatalf("Failed to insert expired secret: %v", err)
	}

	// 3. Run DeleteExpired with 1 day TTL
	deleted, err := DeleteExpired(db, 1)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}

	// 4. Verify results
	if deleted != 1 {
		t.Errorf("Deleted %d records, expected 1", deleted)
	}

	// Fresh secret should remain
	if s, _ := Retrieve(db, freshKey); s == nil {
		t.Error("Fresh secret was incorrectly deleted")
	}

	// Expired secret should be gone
	if s, _ := Retrieve(db, expiredKey); s != nil {
		t.Error("Expired secret was not deleted")
	}
}
