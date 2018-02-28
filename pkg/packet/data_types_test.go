// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func AssertFlags(t *testing.T, encoded byte, decoded flags) {
	t.Helper()
	a := assertions.New(t)
	a.So(newFlags(encoded), should.Equal, decoded)
	a.So(decoded.Byte(), should.Equal, encoded)
}

func TestFlags(t *testing.T) {
	AssertFlags(t, 0x00, flags{false, false, false, false})
	AssertFlags(t, 0x01, flags{true, false, false, false})
	AssertFlags(t, 0x02, flags{false, true, false, false})
	AssertFlags(t, 0x04, flags{false, false, true, false})
	AssertFlags(t, 0x08, flags{false, false, false, true})
}

func AssertStringEncoding(t *testing.T, encoded []byte, decoded string) {
	t.Helper()
	a := assertions.New(t)
	buf := new(bytes.Buffer)
	err := WriteString(buf, decoded)
	a.So(err, should.BeNil)
	a.So(buf.Bytes(), should.Resemble, encoded)
	str, err := ReadString(buf)
	a.So(err, should.BeNil)
	a.So(str, should.Equal, decoded)
}

func TestString(t *testing.T) {
	a := assertions.New(t)
	AssertStringEncoding(t, []byte{0x00, 0x00}, "")
	AssertStringEncoding(t, []byte{0x00, 0x03, 'f', 'o', 'o'}, "foo")
	AssertStringEncoding(t, []byte{0x00, 0x03, 0xEF, 0xBB, 0xBF}, "\U0000FEFF")
	AssertStringEncoding(t, []byte{0x00, 0x05, 'A', 0xF0, 0xAA, 0x9B, 0x94}, "A\U0002A6D4")
	{
		_, err := ReadString(new(bytes.Buffer))
		a.So(err, should.NotBeNil)
	}
	{
		err := WriteString(new(bytes.Buffer), strings.Repeat("foo", 25000))
		a.So(err, should.NotBeNil)
	}
}

func AssertRemainingLengthEncoding(t *testing.T, encoded []byte, decoded int) {
	t.Helper()
	a := assertions.New(t)
	buf := new(bytes.Buffer)
	err := WriteRemainingLength(buf, decoded)
	a.So(err, should.BeNil)
	a.So(buf.Bytes(), should.Resemble, encoded)
	length, err := ReadRemainingLength(buf)
	a.So(err, should.BeNil)
	a.So(length, should.Equal, decoded)
}

func TestRemainingLength(t *testing.T) {
	a := assertions.New(t)
	AssertRemainingLengthEncoding(t, []byte{0x00}, 0)
	AssertRemainingLengthEncoding(t, []byte{0x7F}, 127)
	AssertRemainingLengthEncoding(t, []byte{0x80, 0x01}, 128)
	AssertRemainingLengthEncoding(t, []byte{0xFF, 0x7F}, 16383)
	AssertRemainingLengthEncoding(t, []byte{0x80, 0x80, 0x01}, 16384)
	AssertRemainingLengthEncoding(t, []byte{0xFF, 0xFF, 0x7F}, 2097151)
	AssertRemainingLengthEncoding(t, []byte{0x80, 0x80, 0x80, 0x01}, 2097152)
	AssertRemainingLengthEncoding(t, []byte{0xFF, 0xFF, 0xFF, 0x7F}, 268435455)
	{
		_, err := ReadRemainingLength(new(bytes.Buffer))
		a.So(err, should.NotBeNil)
	}
	{
		err := WriteRemainingLength(new(bytes.Buffer), 300000000)
		a.So(err, should.NotBeNil)
	}
}
