package scrypto

import (
	"slices"
	"testing"
)

func TestPadPKCS5(t *testing.T) {
	tests := []struct {
		data []byte
		want []byte
	}{
		{data: []byte(""), want: []byte("\x04\x04\x04\x04")},
		{data: []byte("a"), want: []byte("a\x03\x03\x03")},
		{data: []byte("ab"), want: []byte("ab\x02\x02")},
		{data: []byte("abc"), want: []byte("abc\x01")},
		{data: []byte("abcd"), want: []byte("abcd\x04\x04\x04\x04")},
		{data: []byte("abcde"), want: []byte("abcde\x03\x03\x03")},
		{data: []byte("abcdefgh"), want: []byte("abcdefgh\x04\x04\x04\x04")},
	}
	for _, tt := range tests {
		got := PadPKCS5(tt.data, 4)
		if slices.Compare(got, tt.want) != 0 {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}

func TestUnpadPKCS5(t *testing.T) {
	tests := []struct {
		data []byte
		want []byte
	}{
		{data: []byte("\x04\x04\x04\x04"), want: []byte("")},
		{data: []byte("a\x03\x03\x03"), want: []byte("a")},
		{data: []byte("ab\x02\x02"), want: []byte("ab")},
		{data: []byte("abc\x01"), want: []byte("abc")},
		{data: []byte("abcd\x04\x04\x04\x04"), want: []byte("abcd")},
		{data: []byte("abcde\x03\x03\x03"), want: []byte("abcde")},
		{data: []byte("abcdefgh\x04\x04\x04\x04"), want: []byte("abcdefgh")},
	}
	for _, tt := range tests {
		got, err := UnpadPKCS5(tt.data, 4)
		if err != nil {
			t.Errorf("got err %+v while wanted %v", err, tt.want)
		} else {
			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		}
	}
}
