package app

import "crypto/md5"

func md5Sum(raw []byte) [16]byte {
	return md5.Sum(raw)
}
