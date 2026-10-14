package utils

import (
	"gamekit/src/base/logger"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gobwas/ws"
)

// Header size length bounds in bytes.
const (
	MaxHeaderSize = 14
	MinHeaderSize = 2
)

// OpCode represents operation code.
type OpCode byte

// Operation codes defined by specification.
// See https://tools.ietf.org/html/rfc6455#section-5.2
const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// IsControl checks whether the c is control operation code.
// See https://tools.ietf.org/html/rfc6455#section-5.5
func (c OpCode) IsControl() bool {
	// RFC6455: Control frames are identified by opcodes where
	// the most significant bit of the opcode is 1.
	//
	// Note that OpCode is only 4 bit length.
	return c&0x8 != 0
}

// IsData checks whether the c is data operation code.
// See https://tools.ietf.org/html/rfc6455#section-5.6
func (c OpCode) IsData() bool {
	// RFC6455: Data frames (e.g., non-control frames) are identified by opcodes
	// where the most significant bit of the opcode is 0.
	//
	// Note that OpCode is only 4 bit length.
	return c&0x8 == 0
}

// IsReserved checks whether the c is reserved operation code.
// See https://tools.ietf.org/html/rfc6455#section-5.2
func (c OpCode) IsReserved() bool {
	// RFC6455:
	// %x3-7 are reserved for further non-control frames
	// %xB-F are reserved for further control frames
	return (0x3 <= c && c <= 0x7) || (0xb <= c && c <= 0xf)
}

var (
	ErrHeaderLengthMSB        = fmt.Errorf("header error: the most significant bit must be 0")
	ErrHeaderLengthUnexpected = fmt.Errorf("header error: unexpected payload length bits")

	// 不处理分帧
	ErrMessageFragmented = fmt.Errorf("not support fragmented process")

	// ErrFrameTooLarge indicates that a message of length higher than
	// MaxFrameSize was being read.
	ErrFrameTooLarge = fmt.Errorf("frame too large")
)

const (
	bit0 = 0x80
	bit1 = 0x40
	bit2 = 0x20
	bit3 = 0x10
	bit4 = 0x08
	bit5 = 0x04
	bit6 = 0x02
	bit7 = 0x01

	len7  = int64(125)
	len16 = int64(^(uint16(0)))
	len64 = int64(^(uint64(0)) >> 1)
)

// WriteHeader writes header binary representation into w.
func serializeHeader(h Header) ([]byte, error) {
	// Make slice of bytes with capacity 14 that could hold any header.
	bts := make([]byte, MaxHeaderSize)

	if h.Fin {
		bts[0] |= bit0
	}
	bts[0] |= h.Rsv << 4
	bts[0] |= byte(h.OpCode)

	var n int
	switch {
	case h.Length <= len7:
		bts[1] = byte(h.Length)
		n = 2

	case h.Length <= len16:
		bts[1] = 126
		binary.BigEndian.PutUint16(bts[2:4], uint16(h.Length))
		n = 4

	case h.Length <= len64:
		bts[1] = 127
		binary.BigEndian.PutUint64(bts[2:10], uint64(h.Length))
		n = 10

	default:
		return bts[:0], ErrHeaderLengthUnexpected
	}

	if h.Masked {
		bts[1] |= bit0
		n += copy(bts[n:], h.Mask[:])
	}

	return bts[:n], nil
}

// Header represents websocket frame header.
// See https://tools.ietf.org/html/rfc6455#section-5.2
type Header struct {
	Fin    bool
	Rsv    byte
	OpCode OpCode
	Masked bool
	Mask   [4]byte
	Length int64
}

// HeaderSize returns number of bytes that are needed to encode given header.
// It returns -1 if header is malformed.
func HeaderSize(h Header) (n int) {
	switch {
	case h.Length < 126:
		n = 2
	case h.Length <= len16:
		n = 4
	case h.Length <= len64:
		n = 10
	default:
		return -1
	}
	if h.Masked {
		n += len(h.Mask)
	}
	return n
}

// gnet.Conn
type IReader interface {
	Peek(n int) (buf []byte, err error)
	Discard(n int) (discarded int, err error)
	InboundBuffered() int
}

// ReadHeader reads a frame header from r.
func ReadHeader(ireader IReader) (ok bool, h Header, err error) {
	buflen := ireader.InboundBuffered()
	// WS 头部至少有两个字节, 根据头部两个字节能确定头部数据
	if buflen <= 2 {
		return
	}
	bts, _ := ireader.Peek(2)

	h.Fin = bts[0]&bit0 != 0
	h.Rsv = (bts[0] & 0x70) >> 4
	h.OpCode = OpCode(bts[0] & 0x0f)

	var extra int

	if bts[1]&bit0 != 0 {
		h.Masked = true
		extra += 4
	}

	length := bts[1] & 0x7f
	switch {
	case length < 126:
		h.Length = int64(length)

	case length == 126:
		extra += 2

	case length == 127:
		extra += 8

	default:
		err = ErrHeaderLengthUnexpected
		return
	}

	if extra == 0 {
		return
	}

	if buflen < (2 + extra) {
		return
	}
	offset := 2 + extra
	bts, _ = ireader.Peek(offset)
	bts = bts[2:]
	switch {
	case length == 126:
		h.Length = int64(binary.BigEndian.Uint16(bts[:2]))
		bts = bts[2:]

	case length == 127:
		if bts[0]&0x80 != 0 {
			err = ErrHeaderLengthMSB
			return
		}
		h.Length = int64(binary.BigEndian.Uint64(bts[:8]))
		bts = bts[8:]
	}

	if h.Masked {
		copy(h.Mask[:], bts)
	}
	// 完整头部
	ok = true
	return
}

// websocket 打包 server -> client binary
func WsServEncode(p []byte) ([]byte, error) {
	op := OpBinary
	header := Header{
		Fin:    true,
		OpCode: op,
		Length: int64(len(p)),
	}

	headerBytes, err := serializeHeader(header)
	if err != nil {
		return headerBytes, err
	}
	lenFrame := len(p) + len(headerBytes)

	buffer := make([]byte, lenFrame)
	copy(buffer, headerBytes)
	copy(buffer[len(headerBytes):], p)

	return buffer, nil
}

func WsEncodePong() ([]byte, error) {
	header := Header{
		Fin:    true,
		OpCode: OpPong,
		Masked: false,
		Mask:   [4]byte{0},
		Length: 0,
	}

	return serializeHeader(header)
}

func WsEncodeClose(reason int) ([]byte, error) {
	header := Header{
		Fin:    true,
		OpCode: OpClose,
		Masked: false,
		Mask:   [4]byte{0},
		Length: 0,
	}

	return serializeHeader(header)
}

var emptyFrame []byte

/***
 * @brief 解ws包
 * @param[in] ireader: 要解包的 reader
 *
 * @retval
 *  bool: 是否是一个完整包
 *  []byte: 消息包缓冲区
 *  int: @see frame.go.OpCode, 直接忽略, 如果不等于 OpBinary 需要打印出来，方便定位问题
 *  bool: true - 分帧，当前不支持客户端分帧(上行数据包很小，没必要分帧)，如果有分帧则直接断开连接
 *  error: 解析报错，报错则直接关闭连接
 */
func WsServDecode(ireader IReader) (bool, []byte, OpCode, bool, error) {
	ok, hdr, err := ReadHeader(ireader)
	// 未满一个包 || 包头解析失败
	if !ok || err != nil {
		return false, emptyFrame, OpBinary, false, err
	}

	// @NOTE: 不检测包头
	// if err = CheckHeader(hdr, StateServerSide); err != nil {
	//     return false, emptyFrame, 0, false, err
	// }

	if hdr.Length >= math.MaxInt32 {
		return false, emptyFrame, hdr.OpCode, false, ErrFrameTooLarge
	}

	offset := HeaderSize(hdr)
	// 未满一个包
	totalLen := offset + int(hdr.Length)
	if totalLen > ireader.InboundBuffered() {
		return false, emptyFrame, hdr.OpCode, false, nil
	}

	pack, _ := ireader.Peek(totalLen)
	pack = pack[offset:]
	if hdr.Masked {
		// 解密
		ws.Cipher(pack, hdr.Mask, 0)
	}
	ireader.Discard(totalLen)

	// NOTE: 不处理 Extensions
	// NOTE: 不处理分帧，但如果有分帧标志需要回调到上层，用于定位问题
	if !hdr.Fin {
		logger.Log.Error("fragmented frame, OpCode: %d, frame len: %d", hdr.OpCode, len(pack))
		return false, emptyFrame, hdr.OpCode, true, ErrMessageFragmented
	}

	// r.opCode = hdr.OpCode
	// NOTE: 不支持 OpText 数据格式，客户端、服务器都是 OpBinary 数据流
	// NOTE: 不处理 OpContinuation
	return true, pack, hdr.OpCode, false, nil
}
