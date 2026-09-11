# go-essentials-pii-aes

Encrypting PII with AES-256-GCM from the standard library: `crypto/aes`, `crypto/cipher`, a random nonce, and authenticated decryption that detects the wrong key.

### How it works

`src/main.go` derives a 256-bit key from a passphrase with SHA-256, encrypts plaintext with AES-GCM using a fresh random nonce (prepended to the ciphertext), round-trips it back to plaintext, and shows that decrypting with the wrong key fails authentication instead of returning garbage.

### Run

```bash
./run.sh
```

### Output

```
plaintext: SSN 123-45-6789
encrypted: fad8ca6957f1a95a97e9d7da3fbaa910704cc515683b75bd...
decrypted: SSN 123-45-6789
roundtrip ok: true
wrong key: cipher: message authentication failed
```
