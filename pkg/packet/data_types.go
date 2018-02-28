// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"encoding/binary"
	"errors"
	"io"
)

func bit(b bool) byte {
	if b {
		return 1
	}
	return 0
}

type flags [8]bool

func newFlags(b byte) (f flags) {
	for i := uint(0); i < 8; i++ {
		f[i] = (b>>i)&1 > 0
	}
	return
}

func (f flags) Byte() (b byte) {
	for i := uint(0); i < 8; i++ {
		b |= bit(f[i]) << i
	}
	return
}

// ErrInvalidFlags is returned if the flags of a packet are invalid
var ErrInvalidFlags = errors.New("Invalid flags")

func validateFlags(expected, got flags) error {
	if expected != got {
		return ErrInvalidFlags
	}
	return nil
}

// WriteByte writes a byte to the given Writer
func WriteByte(w io.Writer, b byte) (err error) {
	buf := make([]byte, 1)
	buf[0] = b
	_, err = w.Write(buf)
	return
}

// ReadByte reads a byte from the given Reader
func ReadByte(r io.Reader) (b byte, err error) {
	buf := make([]byte, 1)
	_, err = r.Read(buf)
	if err != nil {
		return
	}
	b = buf[0]
	return
}

// readFlags reads a flags byte from the given Reader
func readFlags(r io.Reader) (f flags, err error) {
	buf := make([]byte, 1)
	_, err = r.Read(buf)
	if err != nil {
		return
	}
	f = newFlags(buf[0])
	return
}

func encodeUint16(i uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, i)
	return buf
}

// WriteUint16 writes an uint16 to the given Writer
func WriteUint16(w io.Writer, i uint16) (err error) {
	_, err = w.Write(encodeUint16(i))
	return
}

func decodeUint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

// ReadUint16 reads an uint16 from the given Reader
func ReadUint16(r io.Reader) (i uint16, err error) {
	buf := make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		return
	}
	i = decodeUint16(buf)
	return
}

func encodeUint32(i uint32) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint32(buf, i)
	return buf
}

// WriteUint32 writes an uint32 to the given Writer
func WriteUint32(w io.Writer, i uint32) (err error) {
	_, err = w.Write(encodeUint32(i))
	return
}

func decodeUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

// ReadUint32 reads an uint32 from the given Reader
func ReadUint32(r io.Reader) (i uint32, err error) {
	buf := make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		return
	}
	i = decodeUint32(buf)
	return
}

// ErrInvalidLength indicates an invalid length
var ErrInvalidLength = errors.New("Invalid Length")

// WriteBytes writes bytes to the given Writer
func WriteBytes(w io.Writer, b []byte) (err error) {
	if len(b) > 65535 {
		return ErrInvalidLength
	}
	err = WriteUint16(w, uint16(len(b)))
	if err != nil {
		return
	}
	_, err = w.Write(b)
	return
}

// ReadBytes reads bytes form the given Reader
func ReadBytes(r io.Reader) (b []byte, err error) {
	var length uint16
	length, err = ReadUint16(r)
	if err != nil {
		return
	}
	b = make([]byte, length)
	_, err = r.Read(b)
	return
}

// WriteString writes a string to the given Writer
func WriteString(w io.Writer, s string) error {
	return WriteBytes(w, []byte(s))
}

// ReadString reads a string from the given Reader
func ReadString(r io.Reader) (string, error) {
	b, err := ReadBytes(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteStringPair writes a string pair to the given Writer
func WriteStringPair(w io.Writer, k, v string) error {
	if err := WriteBytes(w, []byte(k)); err != nil {
		return err
	}
	return WriteBytes(w, []byte(v))
}

// ReadStringPair reads a string pair from the given Reader
func ReadStringPair(r io.Reader) (string, string, error) {
	k, err := ReadBytes(r)
	if err != nil {
		return "", "", err
	}
	v, err := ReadBytes(r)
	if err != nil {
		return "", "", err
	}
	return string(k), string(v), nil
}

// ErrInvalidRemainingLength when attempting to encode an invalid remaining length field
var ErrInvalidRemainingLength = errors.New("Invalid Remaining Length")

// WriteRemainingLength writes a remaining length field to the given Writer
func WriteRemainingLength(w io.Writer, x int) (err error) {
	var (
		buf         = make([]byte, 0, 4)
		encodedByte byte
	)
	if x > 268435455 {
		return ErrInvalidRemainingLength
	}
	for {
		encodedByte = byte(x % 128)
		x /= 128
		if x > 0 {
			encodedByte |= 128
		}
		buf = append(buf, encodedByte)
		if x <= 0 {
			break
		}
	}
	_, err = w.Write(buf)
	return
}

// ErrMalformedRemainingLength is returned when attempting to read a malformed remaining length
var ErrMalformedRemainingLength = errors.New("Malformed Remaining Length")

// ReadRemainingLength returns the decoded remaining length field
func ReadRemainingLength(r io.Reader) (value int, err error) {
	var (
		multiplier = 1
		buf        = make([]byte, 1)
	)
	for {
		_, err = r.Read(buf)
		if err != nil {
			return
		}
		value += int(buf[0]&127) * multiplier
		if multiplier > 128*128*128 {
			return 0, ErrMalformedRemainingLength
		}
		multiplier *= 128
		if buf[0]&128 == 0 {
			return
		}
	}
}
