package casino

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// fairnessRoll deterministically maps a seed/index pair into [0, max).
// It enables Provably Fair spins when seeds are provided.
func fairnessRoll(serverSeed, clientSeed string, nonce int64, index int, max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	input := fmt.Sprintf("%s|%s|%d|%d", serverSeed, clientSeed, nonce, index)
	sum := sha256.Sum256([]byte(input))
	v := binary.BigEndian.Uint64(sum[0:8])
	return int(v % uint64(max)), nil
}
