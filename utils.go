package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"math"
	"strings"
)

func getMD5(password string, token string) string {
	h := hmac.New(md5.New, []byte(token))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func getSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

const (
	padChar = "="
	alpha   = "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA"
)

func ordAt(msg string, idx int) uint32 {
	if idx < len(msg) {
		return uint32(msg[idx])
	}
	return 0
}

func sencode(msg string, key bool) []uint32 {
	length := len(msg)
	result := []uint32{}
	for i := 0; i < length; i += 4 {
		result = append(result,
			ordAt(msg, i)|
				(ordAt(msg, i+1)<<8)|
				(ordAt(msg, i+2)<<16)|
				(ordAt(msg, i+3)<<24),
		)
	}
	if key {
		result = append(result, uint32(length))
	}
	return result
}

func lencode(msg []uint32, key bool) string {
	length := len(msg)
	totalLen := (length - 1) << 2
	if key {
		last := msg[length-1]
		if last < uint32(totalLen-3) || last > uint32(totalLen) {
			return ""
		}
		totalLen = int(last)
	}
	result := make([]byte, 0, length*4)
	for i := 0; i < length; i++ {
		result = append(result,
			byte(msg[i]&0xFF),
			byte((msg[i]>>8)&0xFF),
			byte((msg[i]>>16)&0xFF),
			byte((msg[i]>>24)&0xFF),
		)
	}
	return string(result[:totalLen])
}

func getXEncode(msg string, key string) string {
	if msg == "" {
		return ""
	}

	pwd := sencode(msg, true)
	pwdk := sencode(key, false)

	for len(pwdk) < 4 {
		pwdk = append(pwdk, 0)
	}

	n := len(pwd) - 1
	if n < 0 {
		return ""
	}

	z := pwd[n]
	y := pwd[0]
	c := uint32(0x86014019 | 0x183639A0)
	var m, e uint32
	var p int
	q := int(math.Floor(6.0 + 52.0/float64(n+1)))
	d := uint32(0)

	for q > 0 {
		d = (d + c) & (0x8CE0D9BF | 0x731F2640)
		e = (d >> 2) & 3
		p = 0
		for p < n {
			y = pwd[p+1]
			m = (z >> 5) ^ (y << 2)
			m = m + ((y >> 3) ^ (z << 4) ^ (d ^ y))
			m = m + (pwdk[(p&3)^int(e)] ^ z)
			pwd[p] = (pwd[p] + m) & (0xEFB8D130 | 0x10472ECF)
			z = pwd[p]
			p++
		}
		y = pwd[0]
		m = (z >> 5) ^ (y << 2)
		m = m + ((y >> 3) ^ (z << 4) ^ (d ^ y))
		m = m + (pwdk[(p&3)^int(e)] ^ z)
		pwd[n] = (pwd[n] + m) & (0xBB390742 | 0x44C6F8BD)
		z = pwd[n]
		q--
	}

	return lencode(pwd, false)
}

func getBase64(s string) string {
	imax := len(s) - len(s)%3

	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	result.Grow((len(s) + 2) / 3 * 4)

	for i := 0; i < imax; i += 3 {
		b10 := (int(s[i]) << 16) | (int(s[i+1]) << 8) | int(s[i+2])
		result.WriteByte(alpha[b10>>18])
		result.WriteByte(alpha[(b10>>12)&63])
		result.WriteByte(alpha[(b10>>6)&63])
		result.WriteByte(alpha[b10&63])
	}

	i := imax
	if len(s)-imax == 1 {
		b10 := int(s[i]) << 16
		result.WriteByte(alpha[b10>>18])
		result.WriteByte(alpha[(b10>>12)&63])
		result.WriteByte(padChar[0])
		result.WriteByte(padChar[0])
	} else if len(s)-imax == 2 {
		b10 := (int(s[i]) << 16) | (int(s[i+1]) << 8)
		result.WriteByte(alpha[b10>>18])
		result.WriteByte(alpha[(b10>>12)&63])
		result.WriteByte(alpha[(b10>>6)&63])
		result.WriteByte(padChar[0])
	}

	return result.String()
}
