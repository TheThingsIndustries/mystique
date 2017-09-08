// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestTypesEncoding(t *testing.T) {
	gunit.Run(new(TypesFixture), t)
}

type TypesFixture struct {
	*gunit.Fixture
}

func (f *TypesFixture) AssertFlags(encoded byte, decoded flags) {
	f.So(newFlags(encoded), should.Equal, decoded)
	f.So(decoded.Byte(), should.Equal, encoded)
}

func (f *TypesFixture) TestFlags() {
	f.AssertFlags(0x00, flags{false, false, false, false})
	f.AssertFlags(0x01, flags{true, false, false, false})
	f.AssertFlags(0x02, flags{false, true, false, false})
	f.AssertFlags(0x04, flags{false, false, true, false})
	f.AssertFlags(0x08, flags{false, false, false, true})
}

func (f *TypesFixture) AssertStringEncoding(encoded []byte, decoded string) {
	buf := new(bytes.Buffer)
	err := WriteString(buf, decoded)
	f.So(err, should.BeNil)
	f.So(buf.Bytes(), should.Resemble, encoded)
	str, err := ReadString(buf)
	f.So(err, should.BeNil)
	f.So(str, should.Equal, decoded)
}

func (f *TypesFixture) TestString() {
	f.AssertStringEncoding([]byte{0x00, 0x00}, "")
	f.AssertStringEncoding([]byte{0x00, 0x03, 'f', 'o', 'o'}, "foo")
	f.AssertStringEncoding([]byte{0x00, 0x03, 0xEF, 0xBB, 0xBF}, "\U0000FEFF")
	f.AssertStringEncoding([]byte{0x00, 0x05, 'A', 0xF0, 0xAA, 0x9B, 0x94}, "A\U0002A6D4")
	{
		_, err := ReadString(new(bytes.Buffer))
		f.So(err, should.NotBeNil)
	}
	{
		err := WriteString(new(bytes.Buffer), strings.Repeat("foo", 25000))
		f.So(err, should.NotBeNil)
	}
}

func (f *TypesFixture) AssertRemainingLengthEncoding(encoded []byte, decoded int) {
	buf := new(bytes.Buffer)
	err := WriteRemainingLength(buf, decoded)
	f.So(err, should.BeNil)
	f.So(buf.Bytes(), should.Resemble, encoded)
	length, err := ReadRemainingLength(buf)
	f.So(err, should.BeNil)
	f.So(length, should.Equal, decoded)
}

func (f *TypesFixture) TestRemainingLength() {
	f.AssertRemainingLengthEncoding([]byte{0x00}, 0)
	f.AssertRemainingLengthEncoding([]byte{0x7F}, 127)
	f.AssertRemainingLengthEncoding([]byte{0x80, 0x01}, 128)
	f.AssertRemainingLengthEncoding([]byte{0xFF, 0x7F}, 16383)
	f.AssertRemainingLengthEncoding([]byte{0x80, 0x80, 0x01}, 16384)
	f.AssertRemainingLengthEncoding([]byte{0xFF, 0xFF, 0x7F}, 2097151)
	f.AssertRemainingLengthEncoding([]byte{0x80, 0x80, 0x80, 0x01}, 2097152)
	f.AssertRemainingLengthEncoding([]byte{0xFF, 0xFF, 0xFF, 0x7F}, 268435455)
	{
		_, err := ReadRemainingLength(new(bytes.Buffer))
		f.So(err, should.NotBeNil)
	}
	{
		err := WriteRemainingLength(new(bytes.Buffer), 300000000)
		f.So(err, should.NotBeNil)
	}
}
