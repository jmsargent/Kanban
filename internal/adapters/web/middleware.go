package web

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// sessionData is the payload stored in the encrypted HttpOnly cookie.
type sessionData struct {
	Token       string `json:"token"`
	DisplayName string `json:"display_name"`
}

// EncryptSession encodes the token and display name into an AES-256-GCM
// encrypted, base64-encoded cookie value. key must be exactly 32 bytes.
func EncryptSession(key []byte, token, displayName string) (string, error) {
	plaintext, err := json.Marshal(sessionData{Token: token, DisplayName: displayName})
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decryptSession decodes and decrypts a cookie value produced by EncryptSession.
// Returns an error if the cookie is missing, malformed, or tampered with.
func decryptSession(key []byte, cookieValue string) (*sessionData, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("decode cookie: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var data sessionData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &data, nil
}

// RequireAuth returns an http.Handler that wraps next with authentication
// enforcement. If the request carries a valid kanban_session cookie (encrypted
// with key), the request is passed to next. Otherwise, the user is redirected
// to GET /auth/token to enter their credentials.
func RequireAuth(key []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("kanban_session")
		if err != nil {
			http.Redirect(w, r, "/auth/token", http.StatusFound)
			return
		}

		if _, err := decryptSession(key, cookie.Value); err != nil {
			log.Printf("auth: rejected tampered or invalid session cookie from %s: %v", r.RemoteAddr, err)
			http.Redirect(w, r, "/auth/token", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
