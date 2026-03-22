package casino

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func GenerateServerSeed() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func HashServerSeed(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(h[:])
}

func Resolve(serverSeed, clientSeed string, nonce int) (roll int, roundHash string) {
	message := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	mac := hmac.New(sha256.New, []byte(serverSeed))
	mac.Write([]byte(message))
	hashBytes := mac.Sum(nil)
	roundHash = hex.EncodeToString(hashBytes)

	// Take first 4 bytes as uint32, mod 100
	num := new(big.Int).SetBytes(hashBytes[:4])
	mod := new(big.Int).SetInt64(100)
	roll = int(new(big.Int).Mod(num, mod).Int64())
	return
}
