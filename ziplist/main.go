// ziplist lists ZIP central-directory entry names as UTF-8 lines on stdout.
//
// Why: macOS Info-ZIP `unzip -Z1` can turn non-ASCII names into `?` when the
 // UTF-8 flag is unset, so joined paths no longer match extracted files.
// This tool reads raw central-directory bytes and decodes names as UTF-8
// (latin1 fallback), matching MBWS plugin attachment `ziplist`.
//
// Usage: ziplist <zipPath>
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	sigEOCD         = []byte{0x50, 0x4b, 0x05, 0x06}
	sigZip64Locator = []byte{0x50, 0x4b, 0x06, 0x07}
	sigCD           = []byte{0x50, 0x4b, 0x01, 0x02}
)

func findLast(buf, sig []byte) int {
	for i := len(buf) - len(sig); i >= 0; i-- {
		ok := true
		for j := 0; j < len(sig); j++ {
			if buf[i+j] != sig[j] {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}

func decodeName(b []byte) string {
	// Prefer UTF-8; fall back to latin1 to preserve bytes.
	if utf8Valid(b) {
		return string(b)
	}
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	return string(runes)
}

func utf8Valid(b []byte) bool {
	i := 0
	for i < len(b) {
		c := b[i]
		if c < 0x80 {
			i++
			continue
		}
		var size int
		switch {
		case c&0xE0 == 0xC0:
			size = 2
		case c&0xF0 == 0xE0:
			size = 3
		case c&0xF8 == 0xF0:
			size = 4
		default:
			return false
		}
		if i+size > len(b) {
			return false
		}
		for j := 1; j < size; j++ {
			if b[i+j]&0xC0 != 0x80 {
				return false
			}
		}
		i += size
	}
	return true
}

func listNames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size < 22 {
		return nil, errors.New("文件过小，不是有效 ZIP")
	}

	tailLen := size
	if tailLen > 66_000 {
		tailLen = 66_000
	}
	tail := make([]byte, tailLen)
	if _, err := f.ReadAt(tail, size-tailLen); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	var cdOffset, cdSize int64 = -1, -1

	if loc := findLast(tail, sigZip64Locator); loc >= 0 && loc+20 <= len(tail) {
		z64eocdOff := int64(binary.LittleEndian.Uint64(tail[loc+8:]))
		if z64eocdOff >= 0 && z64eocdOff+56 <= size {
			z64 := make([]byte, 56)
			if _, err := f.ReadAt(z64, z64eocdOff); err == nil &&
				z64[0] == 0x50 && z64[1] == 0x4b && z64[2] == 0x06 && z64[3] == 0x06 {
				cdSize = int64(binary.LittleEndian.Uint64(z64[40:]))
				cdOffset = int64(binary.LittleEndian.Uint64(z64[48:]))
			}
		}
	}

	if cdOffset < 0 {
		eocd := findLast(tail, sigEOCD)
		if eocd < 0 {
			return nil, errors.New("未找到 ZIP EOCD")
		}
		cdSize = int64(binary.LittleEndian.Uint32(tail[eocd+12:]))
		cdOffset = int64(binary.LittleEndian.Uint32(tail[eocd+16:]))
		if cdOffset == 0xffffffff || cdSize == 0xffffffff {
			return nil, errors.New("ZIP64 中央目录偏移无效，无法枚举")
		}
	}

	if cdOffset < 0 || cdSize <= 0 || cdOffset+cdSize > size {
		return nil, fmt.Errorf("中央目录范围非法 offset=%d size=%d", cdOffset, cdSize)
	}

	cd := make([]byte, cdSize)
	if _, err := f.ReadAt(cd, cdOffset); err != nil {
		return nil, err
	}

	var names []string
	pos := 0
	for pos+46 <= len(cd) {
		if cd[pos] != sigCD[0] || cd[pos+1] != sigCD[1] || cd[pos+2] != sigCD[2] || cd[pos+3] != sigCD[3] {
			break
		}
		fnLen := int(binary.LittleEndian.Uint16(cd[pos+28:]))
		exLen := int(binary.LittleEndian.Uint16(cd[pos+30:]))
		cmLen := int(binary.LittleEndian.Uint16(cd[pos+32:]))
		nameStart := pos + 46
		nameEnd := nameStart + fnLen
		if nameEnd > len(cd) {
			break
		}
		names = append(names, decodeName(cd[nameStart:nameEnd]))
		pos = nameEnd + exLen + cmLen
	}
	return names, nil
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		fmt.Fprintln(os.Stderr, "usage: ziplist <zipPath>")
		os.Exit(2)
	}
	names, err := listNames(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	for _, n := range names {
		fmt.Println(n)
	}
}
