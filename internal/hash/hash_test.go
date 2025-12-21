package hash

import (
	"testing"
)

func TestCalculateHMAC(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		expected string
	}{
		{
			name:     "valid data and key",
			data:     []byte("test message"),
			key:      "secret",
			expected: "b2b0e4e4c2c2c5f0e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1", // будет другим
		},
		{
			name:     "empty data",
			data:     []byte(""),
			key:      "secret",
			expected: "",
		},
		{
			name:     "empty key",
			data:     []byte("test message"),
			key:      "",
			expected: "",
		},
		{
			name:     "different data same key",
			data:     []byte("another message"),
			key:      "secret",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHMAC(tt.data, tt.key)
			if tt.key == "" {
				if result != "" {
					t.Errorf("CalculateHMAC() with empty key = %v, want empty string", result)
				}
			} else if tt.data != nil && len(tt.data) > 0 {
				if result == "" {
					t.Errorf("CalculateHMAC() = empty, want non-empty hash")
				}
				if len(result) != 64 { // SHA256 hex = 64 символа
					t.Errorf("CalculateHMAC() hash length = %d, want 64", len(result))
				}
			}
		})
	}
}

func TestCalculateHMAC_Consistency(t *testing.T) {
	data := []byte("test message")
	key := "secret"

	hash1 := CalculateHMAC(data, key)
	hash2 := CalculateHMAC(data, key)

	if hash1 != hash2 {
		t.Errorf("CalculateHMAC() not consistent: %v != %v", hash1, hash2)
	}
}

func TestVerifyHMAC(t *testing.T) {
	data := []byte("test message")
	key := "secret"
	validHash := CalculateHMAC(data, key)

	tests := []struct {
		name         string
		data         []byte
		receivedHash string
		key          string
		want         bool
	}{
		{
			name:         "valid hash",
			data:         data,
			receivedHash: validHash,
			key:          key,
			want:         true,
		},
		{
			name:         "invalid hash",
			data:         data,
			receivedHash: "invalid",
			key:          key,
			want:         false,
		},
		{
			name:         "wrong key",
			data:         data,
			receivedHash: validHash,
			key:          "wrong",
			want:         false,
		},
		{
			name:         "empty key with empty hash",
			data:         data,
			receivedHash: "",
			key:          "",
			want:         true,
		},
		{
			name:         "empty key with non-empty hash",
			data:         data,
			receivedHash: validHash,
			key:          "",
			want:         false,
		},
		{
			name:         "different data",
			data:         []byte("different message"),
			receivedHash: validHash,
			key:          key,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyHMAC(tt.data, tt.receivedHash, tt.key); got != tt.want {
				t.Errorf("VerifyHMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateHMAC_KnownValue(t *testing.T) {
	// Тест с известным значением для проверки корректности алгоритма
	data := []byte("The quick brown fox jumps over the lazy dog")
	key := "key"
	// Известный HMAC-SHA256 для этих данных
	expected := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"

	result := CalculateHMAC(data, key)
	if result != expected {
		t.Errorf("CalculateHMAC() = %v, want %v", result, expected)
	}
}
