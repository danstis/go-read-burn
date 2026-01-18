package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	key, password, nonce, salt, fullID, err := GenerateID()

	if err != nil {
		t.Fatalf("GenerateID() returned error: %v", err)
	}

	if len(key) != KeyLength {
		t.Errorf("key length = %d, want %d", len(key), KeyLength)
	}

	if len(password) != PasswordLength {
		t.Errorf("password length = %d, want %d", len(password), PasswordLength)
	}

	if len(nonce) != NonceLength {
		t.Errorf("nonce length = %d, want %d", len(nonce), NonceLength)
	}

	if len(salt) != SaltLength {
		t.Errorf("salt length = %d, want %d", len(salt), SaltLength)
	}

	if len(fullID) != FullIDLength {
		t.Errorf("fullID length = %d, want %d", len(fullID), FullIDLength)
	}

	expectedFullID := key + password + nonce + salt
	if fullID != expectedFullID {
		t.Errorf("fullID = %s, want %s", fullID, expectedFullID)
	}

	for i, c := range fullID {
		if !isBase62Char(byte(c)) {
			t.Errorf("fullID contains invalid character at position %d: %c", i, c)
		}
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		_, _, _, _, fullID, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() returned error: %v", err)
		}

		if ids[fullID] {
			t.Errorf("duplicate ID generated: %s", fullID)
		}
		ids[fullID] = true
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name      string
		fullID    string
		wantKey   string
		wantPass  string
		wantNonce string
		wantSalt  string
		wantErr   error
	}{
		{
			name:      "valid ID",
			fullID:    "12345678" + strings.Repeat("a", 32) + strings.Repeat("b", 16) + strings.Repeat("c", 16),
			wantKey:   "12345678",
			wantPass:  strings.Repeat("a", 32),
			wantNonce: strings.Repeat("b", 16),
			wantSalt:  strings.Repeat("c", 16),
			wantErr:   nil,
		},
		{
			name:    "too short",
			fullID:  "short",
			wantErr: ErrInvalidIDLength,
		},
		{
			name:    "too long",
			fullID:  strings.Repeat("a", 73),
			wantErr: ErrInvalidIDLength,
		},
		{
			name:    "invalid characters",
			fullID:  strings.Repeat("!", 72),
			wantErr: ErrInvalidIDCharacters,
		},
		{
			name:    "contains space",
			fullID:  strings.Repeat("a", 35) + " " + strings.Repeat("b", 36),
			wantErr: ErrInvalidIDCharacters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, password, nonce, salt, err := ParseID(tt.fullID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ParseID() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseID() unexpected error: %v", err)
			}

			if key != tt.wantKey {
				t.Errorf("key = %s, want %s", key, tt.wantKey)
			}
			if password != tt.wantPass {
				t.Errorf("password = %s, want %s", password, tt.wantPass)
			}
			if nonce != tt.wantNonce {
				t.Errorf("nonce = %s, want %s", nonce, tt.wantNonce)
			}
			if salt != tt.wantSalt {
				t.Errorf("salt = %s, want %s", salt, tt.wantSalt)
			}
		})
	}
}

func TestParseID_RoundTrip(t *testing.T) {
	key, password, nonce, salt, fullID, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	parsedKey, parsedPassword, parsedNonce, parsedSalt, err := ParseID(fullID)
	if err != nil {
		t.Fatalf("ParseID() error: %v", err)
	}

	if parsedKey != key {
		t.Errorf("parsed key = %s, want %s", parsedKey, key)
	}
	if parsedPassword != password {
		t.Errorf("parsed password = %s, want %s", parsedPassword, password)
	}
	if parsedNonce != nonce {
		t.Errorf("parsed nonce = %s, want %s", parsedNonce, nonce)
	}
	if parsedSalt != salt {
		t.Errorf("parsed salt = %s, want %s", parsedSalt, salt)
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"single byte", "a"},
		{"short text", "hello"},
		{"medium text", "The quick brown fox jumps over the lazy dog"},
		{"with unicode", "Hello, World! - Unicode test"},
		{"exactly 16 bytes", "exactly16bytess!"},
		{"exactly 32 bytes", "exactly32bytesexactly32bytess!!"},
		{"100 bytes", strings.Repeat("x", 100)},
		{"1000 bytes", strings.Repeat("y", 1000)},
		{"4000 bytes", strings.Repeat("z", 4000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, password, nonce, salt, _, err := GenerateID()
			if err != nil {
				t.Fatalf("GenerateID() error: %v", err)
			}

			ciphertext, err := Encrypt(tt.plaintext, password, nonce, salt)
			if err != nil {
				t.Fatalf("Encrypt() error: %v", err)
			}

			decrypted, err := Decrypt(ciphertext, password, nonce, salt)
			if err != nil {
				t.Fatalf("Decrypt() error: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("decrypted = %s, want %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	_, password, nonce, salt, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	_, err = Encrypt("", password, nonce, salt)
	if err != ErrEmptyPlaintext {
		t.Errorf("Encrypt() error = %v, want %v", err, ErrEmptyPlaintext)
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	plaintext := "test secret message"
	ciphertexts := make([][]byte, 10)

	for i := 0; i < 10; i++ {
		_, password, nonce, salt, _, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() error: %v", err)
		}

		ciphertext, err := Encrypt(plaintext, password, nonce, salt)
		if err != nil {
			t.Fatalf("Encrypt() error: %v", err)
		}
		ciphertexts[i] = ciphertext
	}

	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Errorf("encryption produced identical ciphertext for iterations %d and %d", i, j)
			}
		}
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	_, password, nonce, salt, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    bool
	}{
		{"empty", []byte{}, true},
		{"too short for GCM tag", make([]byte, 10), true},
		{"corrupted", make([]byte, 32), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, password, nonce, salt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecrypt_WrongParameters(t *testing.T) {
	_, password1, nonce1, salt1, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	_, password2, nonce2, salt2, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	plaintext := "test secret"
	ciphertext, err := Encrypt(plaintext, password1, nonce1, salt1)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	tests := []struct {
		name     string
		password string
		nonce    string
		salt     string
	}{
		{"wrong password", password2, nonce1, salt1},
		{"wrong nonce", password1, nonce2, salt1},
		{"wrong salt", password1, nonce1, salt2},
		{"all wrong", password2, nonce2, salt2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(ciphertext, tt.password, tt.nonce, tt.salt)
			if err == nil {
				t.Error("Decrypt() should fail with wrong parameters")
			}
		})
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	_, password, nonce, salt, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	plaintext := "test secret"
	ciphertext, err := Encrypt(plaintext, password, nonce, salt)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	_, err = Decrypt(tampered, password, nonce, salt)
	if err != ErrDecryptionFailed {
		t.Errorf("Decrypt() error = %v, want %v", err, ErrDecryptionFailed)
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid all lowercase", strings.Repeat("a", 72), true},
		{"valid all uppercase", strings.Repeat("A", 72), true},
		{"valid all digits", strings.Repeat("0", 72), true},
		{"valid mixed", "12345678" + strings.Repeat("aB", 32), true},
		{"too short", strings.Repeat("a", 71), false},
		{"too long", strings.Repeat("a", 73), false},
		{"empty", "", false},
		{"contains space", strings.Repeat("a", 35) + " " + strings.Repeat("a", 36), false},
		{"contains special char", strings.Repeat("a", 35) + "!" + strings.Repeat("a", 36), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateID(tt.id); got != tt.want {
				t.Errorf("ValidateID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateID_GeneratedID(t *testing.T) {
	for i := 0; i < 100; i++ {
		_, _, _, _, fullID, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() error: %v", err)
		}

		if !ValidateID(fullID) {
			t.Errorf("ValidateID() returned false for generated ID: %s", fullID)
		}
	}
}

func TestBase62Encode(t *testing.T) {
	result := base62Encode([]byte{})
	if result != "" {
		t.Errorf("base62Encode(empty) = %s, want empty string", result)
	}

	result = base62Encode([]byte{0, 0, 0, 1})
	if len(result) == 0 {
		t.Error("base62Encode should produce non-empty result for non-zero input")
	}

	for _, c := range result {
		if !isBase62Char(byte(c)) {
			t.Errorf("base62Encode produced invalid character: %c", c)
		}
	}
}

func TestIsBase62Char(t *testing.T) {
	validChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, c := range validChars {
		if !isBase62Char(byte(c)) {
			t.Errorf("isBase62Char(%c) = false, want true", c)
		}
	}

	invalidChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~ \t\n"
	for _, c := range invalidChars {
		if isBase62Char(byte(c)) {
			t.Errorf("isBase62Char(%c) = true, want false", c)
		}
	}
}

func TestEncrypt_ShortNonce(t *testing.T) {
	_, password, _, salt, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	_, err = Encrypt("test", password, "short", salt)
	if err == nil {
		t.Error("Encrypt() should fail with short nonce")
	}
}

func TestDecrypt_ShortNonce(t *testing.T) {
	_, password, _, salt, _, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error: %v", err)
	}

	_, err = Decrypt(make([]byte, 32), password, "short", salt)
	if err == nil {
		t.Error("Decrypt() should fail with short nonce")
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _, _, err := GenerateID()
		if err != nil {
			b.Fatalf("GenerateID() error: %v", err)
		}
	}
}

func BenchmarkEncrypt(b *testing.B) {
	_, password, nonce, salt, _, _ := GenerateID()
	plaintext := strings.Repeat("x", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(plaintext, password, nonce, salt)
		if err != nil {
			b.Fatalf("Encrypt() error: %v", err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	_, password, nonce, salt, _, _ := GenerateID()
	plaintext := strings.Repeat("x", 1000)
	ciphertext, _ := Encrypt(plaintext, password, nonce, salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decrypt(ciphertext, password, nonce, salt)
		if err != nil {
			b.Fatalf("Decrypt() error: %v", err)
		}
	}
}
