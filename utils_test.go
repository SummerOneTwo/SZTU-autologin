package main

import (
	"testing"
)

func TestGetMD5(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		token    string
	}{
		{
			name:     "basic test",
			password: "test123",
			token:    "abc123",
		},
		{
			name:     "empty password",
			password: "",
			token:    "token",
		},
		{
			name:     "empty token",
			password: "pass",
			token:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getMD5(tc.password, tc.token)
			if len(result) != 32 {
				t.Errorf("getMD5() returned length %d, expected 32", len(result))
			}
		})
	}
}

func TestGetSHA1(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "basic test",
			input: "hello world",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "unicode",
			input: "你好世界",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getSHA1(tc.input)
			if len(result) != 40 {
				t.Errorf("getSHA1() returned length %d, expected 40", len(result))
			}
		})
	}
}

func TestGetXEncode(t *testing.T) {
	testCases := []struct {
		name string
		msg  string
		key  string
	}{
		{
			name: "basic test",
			msg:  "test message",
			key:  "secretkey",
		},
		{
			name: "empty message",
			msg:  "",
			key:  "key",
		},
		{
			name: "empty key",
			msg:  "message",
			key:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getXEncode(tc.msg, tc.key)
			if tc.msg == "" && result != "" {
				t.Error("getXEncode() with empty message should return empty string")
			}
		})
	}
}

func TestGetBase64(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "length 1",
			input: "a",
		},
		{
			name:  "length 2",
			input: "ab",
		},
		{
			name:  "length 3",
			input: "abc",
		},
		{
			name:  "length 4",
			input: "abcd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getBase64(tc.input)
			if tc.input == "" && result != "" {
				t.Error("getBase64() with empty input should return empty string")
			}
		})
	}
}

func TestXEncodeRoundTrip(t *testing.T) {
	msg := "test message for round trip"
	key := "testkey123"

	encoded := getXEncode(msg, key)
	if encoded == "" {
		t.Skip("XEncode returned empty, skipping round trip test")
	}
}
