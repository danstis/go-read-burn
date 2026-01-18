// Package crypto provides AES-256-GCM encryption and decryption functionality
// for secure zero-knowledge secret storage. The server encrypts data but does
// NOT store the decryption key - all encryption parameters are encoded in the
// returned ID which is given only to the user.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/scrypt"
)

const (
	// KeyLength is the length of the database key component in the ID (8 chars).
	KeyLength = 8
	// PasswordLength is the length of the password component in the ID (32 chars).
	PasswordLength = 32
	// NonceLength is the length of the nonce component in the ID (16 chars).
	NonceLength = 16
	// SaltLength is the length of the salt component in the ID (16 chars).
	SaltLength = 16
	// FullIDLength is the total length of a complete ID (72 chars).
	FullIDLength = KeyLength + PasswordLength + NonceLength + SaltLength

	aesKeySize   = 32
	gcmNonceSize = 12

	// scrypt parameters per OWASP recommendations (2^17 minimum for N)
	scryptN = 131072
	scryptR = 8
	scryptP = 1
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	// ErrInvalidIDLength is returned when the ID length is not 72 characters.
	ErrInvalidIDLength = errors.New("invalid ID length: expected 72 characters")
	// ErrInvalidIDCharacters is returned when the ID contains non-base62 characters.
	ErrInvalidIDCharacters = errors.New("invalid ID: contains non-base62 characters")
	// ErrInvalidCiphertext is returned when decryption fails due to invalid data.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrEmptyPlaintext is returned when attempting to encrypt empty data.
	ErrEmptyPlaintext = errors.New("plaintext cannot be empty")
	// ErrDecryptionFailed is returned when authenticated decryption fails.
	ErrDecryptionFailed = errors.New("decryption failed: authentication error")
)

// GenerateID generates a new random ID containing all encryption parameters.
// Returns the individual components and the full 72-character ID.
//
// The ID format is: [8-char Key] + [32-char Password] + [16-char Nonce] + [16-char Salt]
//
//   - key: Used as database lookup key (not secret)
//   - password: Used with scrypt to derive the AES encryption key (secret)
//   - nonce: Nonce for AES-GCM (ensures non-deterministic encryption)
//   - salt: Salt for scrypt key derivation (adds additional randomness)
func GenerateID() (key, password, nonce, salt, fullID string, err error) {
	key, err = generateRandomBase62(KeyLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	password, err = generateRandomBase62(PasswordLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate password: %w", err)
	}

	nonce, err = generateRandomBase62(NonceLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	salt, err = generateRandomBase62(SaltLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	fullID = key + password + nonce + salt
	return key, password, nonce, salt, fullID, nil
}

// ParseID splits a 72-character ID into its components.
// Returns an error if the ID is invalid.
func ParseID(fullID string) (key, password, nonce, salt string, err error) {
	if len(fullID) != FullIDLength {
		return "", "", "", "", ErrInvalidIDLength
	}

	for _, c := range fullID {
		if !isBase62Char(byte(c)) {
			return "", "", "", "", ErrInvalidIDCharacters
		}
	}

	key = fullID[0:KeyLength]
	password = fullID[KeyLength : KeyLength+PasswordLength]
	nonce = fullID[KeyLength+PasswordLength : KeyLength+PasswordLength+NonceLength]
	salt = fullID[KeyLength+PasswordLength+NonceLength : FullIDLength]

	return key, password, nonce, salt, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with the given password, nonce, and salt.
// The password and salt are used with scrypt to derive a 32-byte AES key.
// AES-GCM provides authenticated encryption (confidentiality + integrity).
func Encrypt(plaintext, password, nonce, salt string) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, ErrEmptyPlaintext
	}

	aesKey, err := deriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceBytes := []byte(nonce)
	if len(nonceBytes) < gcmNonceSize {
		return nil, fmt.Errorf("nonce too short: got %d bytes, need %d", len(nonceBytes), gcmNonceSize)
	}
	nonceBytes = nonceBytes[:gcmNonceSize]

	ciphertext := aesGCM.Seal(nil, nonceBytes, []byte(plaintext), nil)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given password, nonce, and salt.
// The password and salt are used with scrypt to derive the AES key.
// Returns an error if authentication fails (tampered ciphertext).
func Decrypt(ciphertext []byte, password, nonce, salt string) (string, error) {
	if len(ciphertext) == 0 {
		return "", ErrInvalidCiphertext
	}

	aesKey, err := deriveKey(password, salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceBytes := []byte(nonce)
	if len(nonceBytes) < gcmNonceSize {
		return "", fmt.Errorf("nonce too short: got %d bytes, need %d", len(nonceBytes), gcmNonceSize)
	}
	nonceBytes = nonceBytes[:gcmNonceSize]

	plaintext, err := aesGCM.Open(nil, nonceBytes, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

func deriveKey(password, salt string) ([]byte, error) {
	return scrypt.Key([]byte(password), []byte(salt), scryptN, scryptR, scryptP, aesKeySize)
}

func generateRandomBase62(length int) (string, error) {
	numBytes := length + 8

	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encoded := base62Encode(randomBytes)

	if len(encoded) < length {
		return generateRandomBase62(length)
	}

	return encoded[:length], nil
}

func base62Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	num := new(big.Int).SetBytes(input)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)

	var result []byte

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base62Alphabet[mod.Int64()])
	}

	for _, b := range input {
		if b == 0 {
			result = append(result, base62Alphabet[0])
		} else {
			break
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func isBase62Char(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

// ValidateID validates that an ID is well-formed (correct length and characters).
func ValidateID(id string) bool {
	if len(id) != FullIDLength {
		return false
	}
	for _, c := range id {
		if !isBase62Char(byte(c)) {
			return false
		}
	}
	return true
}
