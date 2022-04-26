package utils

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/spaolacci/murmur3"
)

// GenMurmur 默默哈希
func GenMurmur(s string) uint32 {
	m := murmur3.New32()
	b := []byte(s)
	m.Write(b)
	return m.Sum32()
}

// RandomNum 指定区间随机数
func RandomNum(min, max int) int {
	rand.Seed(time.Now().UnixNano())

	if min >= max || max == 0 {
		return max
	}

	return rand.Intn(max-min) + min
}

// Int64ToBytes int64转[]byte
func Int64ToBytes(n int64) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, n)
	return buf.Bytes()
}

// SpliceString 拼接字符串
func SpliceString(strs ...string) string {
	var buffer bytes.Buffer
	for i := range strs {
		buffer.WriteString(strs[i])
	}
	return buffer.String()
}
