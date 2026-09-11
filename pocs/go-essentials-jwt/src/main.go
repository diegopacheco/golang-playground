package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func b64Encode(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func sign(payload map[string]any, secret string) string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)
	signingInput := b64Encode(headerJSON) + "." + b64Encode(payloadJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return signingInput + "." + b64Encode(mac.Sum(nil))
}

func verify(token, secret string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expected := b64Encode(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid signature")
	}
	raw, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var payload map[string]any
	json.Unmarshal(raw, &payload)
	if exp, ok := payload["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return payload, nil
}

func main() {
	secret := "super-secret-key"
	payload := map[string]any{"sub": "alice", "role": "admin", "exp": time.Now().Unix() + 3600}

	token := sign(payload, secret)
	fmt.Println("token:", token[:60]+"...")

	verified, _ := verify(token, secret)
	fmt.Printf("verified payload: sub=%v role=%v\n", verified["sub"], verified["role"])

	if _, err := verify(token, "wrong-secret"); err != nil {
		fmt.Println("tamper check:", err)
	}

	expired := sign(map[string]any{"sub": "bob", "exp": time.Now().Unix() - 10}, secret)
	if _, err := verify(expired, secret); err != nil {
		fmt.Println("expiry check:", err)
	}
}
