// Modified version of mail.ru easyjson pool
package buffer

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"
)

const (
	startSize  = 256
	pooledSize = 512
	maxSize    = 32768
	multiply   = 2
	initial    = 8
	two        = 2

	maxArraySize = uint((uint64(1)<<50 - 1) & uint64(^uint(0)>>1))
)

type PoolConfig struct {
	StartSize  int // Minimum chunk size that is allocated.
	PooledSize int // Minimum chunk size that is reused, reusing chunks too small will result in overhead.
	MaxSize    int // Maximum chunk size that will be allocated.
}

// //nolint:gochecknoglobals
var config = PoolConfig{
	StartSize:  startSize,
	PooledSize: pooledSize,
	MaxSize:    maxSize,
}

var errNegativeRead = errors.New("buffer.Buffer: reader returned negative count from Read")

// Reuse pool: chunk size -> pool.
// //nolint:gochecknoglobals
var buffers = map[int]*sync.Pool{}

func initBuffers() {
	for l := config.PooledSize; l <= config.MaxSize; l *= 2 {
		buffers[l] = new(sync.Pool)
	}
}

// //nolint:gochecknoinits
func init() {
	initBuffers()
}

// Init sets up a non-default pooling and allocation strategy. Should be run before serialization is done.
func Init(cfg PoolConfig) {
	config = cfg

	initBuffers()
}

// PutBuf puts a chunk to reuse pool if it can be reused.
func PutBuf(buf []byte) {
	size := cap(buf)

	if size < config.PooledSize {
		return
	}

	if c := buffers[size]; c != nil {
		// Save un unsafe pointer to the array instead of the slice to
		// avoid an extra allocation.
		// We don't care about the length and we know the capability so
		c.Put(unsafe.Pointer(&buf[:1][0]))
	}
}

// getBuf gets a chunk from reuse pool or creates a new one if reuse failed.
func getBuf(size int) []byte {
	if size < config.PooledSize {
		return make([]byte, 0, size)
	}

	if c := buffers[size]; c != nil {
		v := c.Get()
		if v != nil {
			// Recreate back the original slice.
			// Get back the array and add length and capability.
			// Limiting the array to the proper capability will make this
			// safe.
			buf := (*[maxArraySize]byte)(v.(unsafe.Pointer)) //nolint:errcheck,forcetypeassert

			return buf[:0:size]
		}
	}

	return make([]byte, 0, size)
}

// Buffer is a buffer optimized for serialization without extra copying.
type Buffer struct {
	// Buf is the current chunk that can be used for serialization.
	Buf []byte

	toPool []byte
	buffs  [][]byte
}

func (b *Buffer) ReadAll(r io.Reader, buffSize, maxSize int) (int64, error) {
	b.EnsureSpace(maxSize)

	return b.read(r, buffSize)
}

func (b *Buffer) read(r io.Reader, sz int) (int64, error) {
	buf := getBuf(sz)

	n := int64(0)

	for {
		read, readErr := r.Read(buf[len(buf):cap(buf)])
		if read < 0 {
			panic(errNegativeRead)
		}

		if read > 0 {
			b.AppendBytes(buf[:read])
		}

		n += int64(read)

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				PutBuf(buf)

				return n, nil // readErr is EOF, so return nil explicitly
			}

			PutBuf(buf)

			b.clear()

			return n, fmt.Errorf("read error: %w", readErr)
		}

		buf = buf[:0]
	}
}

func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	return b.read(r, config.PooledSize*multiply)
}

// EnsureSpace makes sure that the current chunk contains at least s free bytes,
// possibly creating a new chunk.
//
// //nolint:nestif
func (b *Buffer) EnsureSpace(space int) {
	if cap(b.Buf)-len(b.Buf) >= space {
		return
	}

	length := len(b.Buf)

	if length > 0 {
		if cap(b.toPool) != cap(b.Buf) {
			// Chunk was reallocated, toPool can be pooled.
			PutBuf(b.toPool)
		}

		if cap(b.buffs) == 0 {
			b.buffs = make([][]byte, 0, initial)
		}

		b.buffs = append(b.buffs, b.Buf)
		length = cap(b.toPool) * multiply
	} else {
		if space > config.MaxSize {
			length = config.MaxSize
		} else {
			length = config.PooledSize

			if length < space {
				for length < space && length < config.MaxSize {
					length *= 2
				}
			}
		}
	}

	if length > config.MaxSize {
		length = config.MaxSize
	}

	b.Buf = getBuf(length)
	b.toPool = b.Buf
}

// AppendByte appends a single byte to buffer.
func (b *Buffer) AppendByte(data byte) {
	if cap(b.Buf) == len(b.Buf) {
		b.EnsureSpace(1)
	}

	b.Buf = append(b.Buf, data)
}

func (b *Buffer) AppendTwoBytes(data1, data2 byte) {
	if cap(b.Buf)-len(b.Buf) < two {
		b.EnsureSpace(two)
	}

	b.Buf = append(b.Buf, data1, data2)
}

// Implements io.Writer.
func (b *Buffer) Write(p []byte) (int, error) {
	n := len(p)

	b.AppendBytes(p)

	return n, nil
}

// AppendBytes appends a byte slice to buffer.
func (b *Buffer) AppendBytes(data []byte) {
	if len(data) <= cap(b.Buf)-len(b.Buf) {
		b.Buf = append(b.Buf, data...) // fast path
	} else {
		b.appendBytesSlow(data)
	}
}

func (b *Buffer) appendBytesSlow(data []byte) {
	d := data

	for len(d) > 0 {
		b.EnsureSpace(1)

		sz := cap(b.Buf) - len(b.Buf)
		if sz > len(d) {
			sz = len(d)
		}

		b.Buf = append(b.Buf, d[:sz]...)

		d = d[sz:]
	}
}

// AppendString appends a string to buffer.
func (b *Buffer) AppendString(data string) {
	if len(data) <= cap(b.Buf)-len(b.Buf) {
		b.Buf = append(b.Buf, data...) // fast path
	} else {
		b.appendStringSlow(data)
	}
}

func (b *Buffer) appendStringSlow(s string) {
	data := s

	// //nolint:gocritic
	for len(data) > 0 {
		b.EnsureSpace(1)

		sz := cap(b.Buf) - len(b.Buf)
		if sz > len(data) {
			sz = len(data)
		}

		b.Buf = append(b.Buf, data[:sz]...)
		data = data[sz:]
	}
}

func (b *Buffer) AppendStrings(data ...string) {
	for _, str := range data {
		b.AppendString(str)
	}
}

func (b *Buffer) AppendXMLElement(tag, value string) {
	b.AppendString("<")
	b.AppendString(tag)
	b.AppendString(">")

	b.AppendXMLEncode(value)

	b.AppendString("</")
	b.AppendString(tag)
	b.AppendString(">")
}

func (b *Buffer) AppendXMLEncode(data string) {
	for i := range len(data) {
		char := data[i]
		switch char {
		case '<':
			b.AppendString("&lt;")
		case '>':
			b.AppendString("&gt;")
		case '"':
			b.AppendString("&quot;")
		case '\'':
			b.AppendString("&apos;")
		case '&':
			b.AppendString("&amp;")
		case '\t':
			b.AppendString("&#x9;")
		default:
			b.AppendByte(char)
		}
	}
}

// Size computes the size of a buffer by adding sizes of every chunk.
func (b *Buffer) Size() int {
	size := len(b.Buf)

	for _, buf := range b.buffs {
		size += len(buf)
	}

	return size
}

// WriteTo outputs the contents of a buffer to a writer and resets the buffer.
func (b *Buffer) WriteTo(writer io.Writer) (int64, error) {
	var n int

	var err error

	written := int64(0)

	for _, buf := range b.buffs {
		if err == nil {
			n, err = writer.Write(buf)
			written += int64(n)
		}

		PutBuf(buf)
	}

	if err == nil {
		n, err = writer.Write(b.Buf)
		written += int64(n)
	}

	PutBuf(b.toPool)

	b.buffs = b.buffs[:0]
	b.Buf = nil
	b.toPool = nil

	return written, err
}

// BuildBytes creates a single byte slice with all the contents of the buffer. Data is
// copied if it does not fit in a single chunk. You can optionally provide one byte
// slice as argument that it will try to reuse.
func (b *Buffer) BuildBytes(reuse ...[]byte) []byte {
	if len(b.buffs) == 0 {
		ret := b.Buf
		b.toPool = nil
		b.Buf = nil

		return ret
	}

	var ret []byte

	size := b.Size()

	// If we got a buffer as argument and it is big enough, reuse it.
	if len(reuse) == 1 && cap(reuse[0]) >= size {
		ret = reuse[0][:0]
	} else {
		ret = make([]byte, 0, size)
	}

	for _, buf := range b.buffs {
		ret = append(ret, buf...)
		PutBuf(buf)
	}

	ret = append(ret, b.Buf...)
	PutBuf(b.toPool)

	b.buffs = b.buffs[:0]
	b.toPool = nil
	b.Buf = nil

	return ret
}

type readCloser struct {
	buffs  [][]byte
	offset int
}

func (r *readCloser) Read(input []byte) (int, error) {
	n := 0

	for _, buf := range r.buffs {
		// Copy as much as we can.
		copied := copy(input[n:], buf[r.offset:])
		n += copied // Increment how much we filled.

		// Did we empty the whole buffer?
		if r.offset+copied == len(buf) {
			// On to the next buffer.
			r.offset = 0
			r.buffs = r.buffs[1:]

			// We can release this buffer.
			PutBuf(buf)
		} else {
			r.offset += copied
		}

		if n == len(input) {
			break
		}
	}
	// No buffers left or nothing read?
	if len(r.buffs) == 0 {
		return n, io.EOF
	}

	return n, nil
}

func (r *readCloser) Close() error {
	// Release all remaining buffers.
	for _, buf := range r.buffs {
		PutBuf(buf)
	}
	// In case Close gets called multiple times.
	r.buffs = nil

	return nil
}

func (b *Buffer) clear() {
	for _, buf := range b.buffs {
		PutBuf(buf)
	}

	PutBuf(b.toPool)

	b.toPool = nil
	b.Buf = nil
	b.buffs = nil
}

// ReadCloser creates an io.ReadCloser with all the contents of the buffer.
func (b *Buffer) ReadCloser() io.ReadCloser {
	ret := &readCloser{append(b.buffs, b.Buf), 0}

	b.buffs = nil
	b.toPool = nil
	b.Buf = nil

	return ret
}
