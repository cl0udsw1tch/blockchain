package t_util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"
)


func SliceCompare(a, b []byte) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}


func CopySlice2D(src [][]byte) [][]byte {
	dest := make([][]byte, len(src))
    for i := range src {
        dest[i] = make([]byte, len(src[i]))
        copy(dest[i], src[i]) 
    }
	return dest
}

func Hash256(src []byte) []byte {

	f := sha256.Sum256(src)
	h := sha256.Sum256(f[:])
	return h[:]
}

func ByteLen(val int32) int {
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.BigEndian, val)
	return len(buffer.Bytes())
}

func ConvertToHexString(data [][]byte) string {
	
	var sb strings.Builder
	
	for _, byteSlice := range data {
		hexStr := hex.EncodeToString(byteSlice)
		sb.WriteString(hexStr)
	}

	return sb.String()
}