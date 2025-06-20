/*

CountWriter implementation.

*/

package bitio

import (
	"io"
)

// CountWriter is an improved version of Writer that also keeps track
// of the number of processed bits. If you don't need the number
// of processed bits, use the faster Writer.
//
// For convenience, it also implements io.WriterCloser and io.ByteWriter.
type CountWriter struct {
	Writer
	BitsCount int64 // Total number of bits written
}

// NewCountWriter returns a new CountWriter using the specified io.Writer as the
// output.
//
// Must be closed in order to flush cached data.
// If you can't or don't want to close it, flushing data can also be forced
// by calling Align().
func NewCountWriter(out io.Writer) *CountWriter {
	return &CountWriter{NewWriter(out), 0}
}

// Write writes len(p) bytes (8 * len(p) bits) to the underlying writer.
//
// Write implements io.Writer, and gives a byte-level interface to the bit stream.
// This will give best performance if the underlying io.Writer is aligned
// to a byte boundary (else all the individual bytes are spread to multiple bytes).
// Byte boundary can be ensured by calling Align().
func (w *CountWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.BitsCount += int64(n) * 8
	return
}

// WriteBits writes out the n lowest bits of r.
// Bits of r in positions higher than n are ignored.
//
// For example:
//
//	err := w.WriteBits(0x1234, 8)
//
// is equivalent to:
//
//	err := w.WriteBits(0x34, 8)
func (w *CountWriter) WriteBits(r uint64, n uint8) (err error) {
	// if r would have bits set at n or higher positions (zero indexed),
	// WriteBitsUnsafe's implementation could "corrupt" bits in cache.
	// That is not acceptable. To be on the safe side, mask out higher bits:
	return w.WriteBitsUnsafe((r & (1<<n - 1)), n)
}

// WriteBitsUnsafe writes out the n lowest bits of r.
//
// r must not have bits set at n or higher positions (zero indexed).
// If r might not satisfy this, a mask must be explicitly applied
// before passing it to WriteBitsUnsafe(), or WriteBits() should be used instead.
//
// WriteBitsUnsafe() offers slightly better performance than WriteBits() because
// the input r is not masked. Calling WriteBitsUnsafe() with an r that does
// not satisfy this is undefined behavior (might corrupt previously written bits).
//
// E.g. if you want to write 8 bits:
//
//	err := w.WriteBitsUnsafe(0x34, 8)        // This is OK,
//	                                         // 0x34 has no bits set higher than the 8th
//	err := w.WriteBitsUnsafe(0x1234&0xff, 8) // &0xff masks out bits higher than the 8th
//
// Or:
//
//	err := w.WriteBits(0x1234, 8)            // bits higher than the 8th are ignored here
func (w *CountWriter) WriteBitsUnsafe(r uint64, n uint8) (err error) {
	err = w.Writer.WriteBitsUnsafe(r, n)
	if err == nil {
		w.BitsCount += int64(n)
	}
	return
}

// WriteByte writes 8 bits.
//
// WriteByte implements io.ByteWriter.
func (w *CountWriter) WriteByte(b byte) (err error) {
	err = w.Writer.WriteByte(b)
	if err == nil {
		w.BitsCount += 8
	}
	return
}

// WriteBool writes one bit: 1 if param is true, 0 otherwise.
func (w *CountWriter) WriteBool(b bool) (err error) {
	err = w.Writer.WriteBool(b)
	if err == nil {
		w.BitsCount += 1
	}
	return
}

// Align aligns the bit stream to a byte boundary,
// so next write will start/go into a new byte.
// If there are cached bits, they are first written to the output.
// Returns the number of skipped (unset but still written) bits.
func (w *CountWriter) Align() (skipped uint8, err error) {
	skipped, err = w.Writer.Align()
	w.BitsCount += int64(skipped)
	return
}

// Close closes the bit writer, writes out cached bits.
// It does not close the underlying io.Writer.
//
// Close implements io.Closer.
func (w *CountWriter) Flush() (err error) {
	// Make sure cached bits are flushed:
	if _, err = w.Align(); err != nil {
		return
	}

	return nil
}
