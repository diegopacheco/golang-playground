# go-essentials-jwt

JWT (HS256) built with only the standard library: `crypto/hmac`, `crypto/sha256`, `encoding/base64`, `encoding/json`. No external dependency.

### How it works

`src/main.go` signs a payload into a `header.payload.signature` token, verifies the signature with `hmac.Equal` (constant time), and checks the `exp` claim. It then proves that a wrong secret and an expired token are both rejected.

### Run

```bash
./run.sh
```

### Output

```
token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODM4ODg1ODU...
verified payload: sub=alice role=admin
tamper check: invalid signature
expiry check: token expired
```
