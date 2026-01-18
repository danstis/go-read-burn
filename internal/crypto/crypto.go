// Package crypto provides AES-256-CBC encryption and decryption functionality
// for secure zero-knowledge secret storage. The server encrypts data but does
// NOT store the decryption key - all encryption parameters are encoded in the
// returned ID which is given only to the user.
package crypto

import (
	"bytes"
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
	// IVLength is the length of the initialization vector component in the ID (16 chars).
	IVLength = 16
	// SaltLength is the length of the salt component in the ID (16 chars).
	SaltLength = 16
	// FullIDLength is the total length of a complete ID (72 chars).
	FullIDLength = KeyLength + PasswordLength + IVLength + SaltLength

	aesBlockSize = 16
	aesKeySize   = 32

	// scrypt parameters per OWASP recommendations
	scryptN = 32768
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
	// ErrInvalidPadding is returned when PKCS7 unpadding fails.
	ErrInvalidPadding = errors.New("invalid PKCS7 padding")
	// ErrEmptyPlaintext is returned when attempting to encrypt empty data.
	ErrEmptyPlaintext = errors.New("plaintext cannot be empty")
)

// GenerateID generates a new random ID containing all encryption parameters.
// Returns the individual components and the full 72-character ID.
//
// The ID format is: [8-char Key] + [32-char Password] + [16-char IV] + [16-char Salt]
//
//   - key: Used as database lookup key (not secret)
//   - password: Used with scrypt to derive the AES encryption key (secret)
//   - iv: Initialization vector for AES-CBC (ensures non-deterministic encryption)
//   - salt: Salt for scrypt key derivation (adds additional randomness)
func GenerateID() (key, password, iv, salt, fullID string, err error) {
	key, err = generateRandomBase62(KeyLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	password, err = generateRandomBase62(PasswordLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate password: %w", err)
	}

	iv, err = generateRandomBase62(IVLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate IV: %w", err)
	}

	salt, err = generateRandomBase62(SaltLength)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	fullID = key + password + iv + salt
	return key, password, iv, salt, fullID, nil
}

// ParseID splits a 72-character ID into its components.
// Returns an error if the ID is invalid.
func ParseID(fullID string) (key, password, iv, salt string, err error) {
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
	iv = fullID[KeyLength+PasswordLength : KeyLength+PasswordLength+IVLength]
	salt = fullID[KeyLength+PasswordLength+IVLength : FullIDLength]

	return key, password, iv, salt, nil
}

// Encrypt encrypts plaintext using AES-256-CBC with the given password, iv, and salt.
// The password and salt are used with scrypt to derive a 32-byte AES key.
// The plaintext is padded using PKCS7 before encryption.
func Encrypt(plaintext, password, iv, salt string) ([]byte, error) {
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

	ivBytes := []byte(iv)
	if len(ivBytes) < aesBlockSize {
		return nil, fmt.Errorf("IV too short: got %d bytes, need %d", len(ivBytes), aesBlockSize)
	}
	ivBytes = ivBytes[:aesBlockSize]

	paddedPlaintext := pkcs7Pad([]byte(plaintext), aesBlockSize)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-CBC with the given password, iv, and salt.
// The password and salt are used with scrypt to derive the AES key.
// PKCS7 padding is removed after decryption.
func Decrypt(ciphertext []byte, password, iv, salt string) (string, error) {
	if len(ciphertext) == 0 {
		return "", ErrInvalidCiphertext
	}

	if len(ciphertext)%aesBlockSize != 0 {
		return "", fmt.Errorf("%w: ciphertext length must be multiple of %d", ErrInvalidCiphertext, aesBlockSize)
	}

	aesKey, err := deriveKey(password, salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	ivBytes := []byte(iv)
	if len(ivBytes) < aesBlockSize {
		return "", fmt.Errorf("IV too short: got %d bytes, need %d", len(ivBytes), aesBlockSize)
	}
	ivBytes = ivBytes[:aesBlockSize]

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(plaintext, ciphertext)

	unpaddedPlaintext, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", err
	}

	return string(unpaddedPlaintext), nil
}

func deriveKey(password, salt string) ([]byte, error) {
	return scrypt.Key([]byte(password), []byte(salt), scryptN, scryptR, scryptP, aesKeySize)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padBytes...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPadding
	}

	paddingLen := int(data[len(data)-1])

	if paddingLen == 0 || paddingLen > aesBlockSize {
		return nil, ErrInvalidPadding
	}

	if paddingLen > len(data) {
		return nil, ErrInvalidPadding
	}

	for i := len(data) - paddingLen; i < len(data); i++ {
		if data[i] != byte(paddingLen) {
			return nil, ErrInvalidPadding
		}
	}

	return data[:len(data)-paddingLen], nil
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
