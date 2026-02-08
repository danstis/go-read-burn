package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return nil
	})
	if err != nil {
		t.Errorf("Bucket '%s' was not created", BucketName)
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
	if err := Store(db, key, data); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

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
	if err := Store(db, freshKey, []byte("data")); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// 2. Manually store an expired secret (2 days old)
	expiredKey := "expired"
	expiredSecret := Secret{
		Timestamp: time.Now().Add(-48 * time.Hour).UnixMilli(),
		Encrypted: "expireddata",
	}
	expiredData, _ := json.Marshal(expiredSecret)
	
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
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

func TestDeleteExpired_NoBucket(t *testing.T) {
	db := setupTestDB(t)
	// Do NOT call InitBucket(db)

	_, err := DeleteExpired(db, 1)
	if err == nil {
		t.Error("DeleteExpired should return error when bucket is missing")
	} else if err.Error() != "bucket not found" {
		t.Errorf("Expected 'bucket not found' error, got: %v", err)
	}
}

func TestStore_NoBucket(t *testing.T) {
	db := setupTestDB(t)
	// Do NOT call InitBucket(db)

	err := Store(db, "key", []byte("data"))
	if err == nil {
		t.Error("Store should return error when bucket is missing")
	} else if err.Error() != "bucket not found" {
		t.Errorf("Expected 'bucket not found' error, got: %v", err)
	}
}

func TestRetrieve_NoBucket(t *testing.T) {
	db := setupTestDB(t)
	// Do NOT call InitBucket(db)

	_, err := Retrieve(db, "key")
	if err == nil {
		t.Error("Retrieve should return error when bucket is missing")
	} else if err.Error() != "bucket not found" {
		t.Errorf("Expected 'bucket not found' error, got: %v", err)
	}
}

func TestDelete_NoBucket(t *testing.T) {
	db := setupTestDB(t)
	// Do NOT call InitBucket(db)

	err := Delete(db, "key")
	if err == nil {
		t.Error("Delete should return error when bucket is missing")
	} else if err.Error() != "bucket not found" {
		t.Errorf("Expected 'bucket not found' error, got: %v", err)
	}
}

func TestInitBucket_Idempotent(t *testing.T) {
	db := setupTestDB(t)

	// Call InitBucket multiple times - should not error
	for i := 0; i < 3; i++ {
		if err := InitBucket(db); err != nil {
			t.Fatalf("InitBucket call %d failed: %v", i+1, err)
		}
	}
}

func TestStore_OverwriteExisting(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	key := "overwrite-key"
	originalData := []byte("original")
	newData := []byte("updated")

	// Store original
	if err := Store(db, key, originalData); err != nil {
		t.Fatalf("Store original failed: %v", err)
	}

	// Overwrite with new data
	if err := Store(db, key, newData); err != nil {
		t.Fatalf("Store overwrite failed: %v", err)
	}

	// Retrieve and verify it's the new data
	secret, err := Retrieve(db, key)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if secret == nil {
		t.Fatal("Retrieve returned nil")
	}

	decoded, err := base64.StdEncoding.DecodeString(secret.Encrypted)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if string(decoded) != string(newData) {
		t.Errorf("Got %s, want %s", string(decoded), string(newData))
	}
}

func TestDeleteExpired_NoExpiredSecrets(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	// Store only fresh secrets
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("fresh-%d", i)
		if err := Store(db, key, []byte("data")); err != nil {
			t.Fatalf("Store failed: %v", err)
		}
	}

	// Run DeleteExpired - should delete nothing
	deleted, err := DeleteExpired(db, 1)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Deleted %d records, expected 0", deleted)
	}
}

func TestDeleteExpired_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	// Store invalid JSON directly in bucket
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		return b.Put([]byte("invalid"), []byte("not-json"))
	})
	if err != nil {
		t.Fatalf("Failed to insert invalid data: %v", err)
	}

	// DeleteExpired should skip invalid JSON entries without error
	deleted, err := DeleteExpired(db, 1)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	// Invalid JSON entry should not be deleted (it's skipped)
	if deleted != 0 {
		t.Errorf("Deleted %d records, expected 0", deleted)
	}
}

func TestDelete_NonExistentKey(t *testing.T) {
	db := setupTestDB(t)
	if err := InitBucket(db); err != nil {
		t.Fatalf("Failed to init bucket: %v", err)
	}

	// Delete a key that doesn't exist - should not error
	err := Delete(db, "nonexistent-key")
	if err != nil {
		t.Errorf("Delete returned error for non-existent key: %v", err)
	}
}
