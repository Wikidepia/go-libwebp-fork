package webp

import (
	"fmt"
	"image"
	"io"
	"runtime"
	"unsafe"

	"git.sr.ht/~jackmordaunt/go-libwebp/lib"
	"modernc.org/libc"
)

// Encode an image into webp with default settings.
func Encode(w io.Writer, m image.Image, opt ...EncodeOption) error {
	var enc Encoder
	for _, op := range opt {
		op(&enc)
	}
	return enc.Encode(w, m)
}

// EncodeOption configures the encoder.
type EncodeOption func(*Encoder)

// Quality in the range (0,1].
func Quality(q float32) EncodeOption {
	return func(enc *Encoder) {
		enc.Quality = q
	}
}

// Lossless will ignore quality.
func Lossless() EncodeOption {
	return func(enc *Encoder) {
		enc.Lossless = true
	}
}

// Encoder implements webp encoding of an image.
type Encoder struct {
	// Quality is in the range (0,1]. Values outside of this
	// range will be treated as 1.
	Quality float32
	// Lossless indicates whether to use the lossless compression
	// strategy. If true, the Quality field is ignored.
	Lossless bool
}

// Encode specified image as webp to w.
func (enc *Encoder) Encode(w io.Writer, m image.Image) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if enc.Quality <= 0.0 || enc.Quality > 1 {
		enc.Quality = 1.0
	}
	rgbaImage := image.NewRGBA(m.Bounds())
	rect := m.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			rgbaImage.Set(x, y, m.At(x, y))
		}
	}
	var (
		// out buffer to contain webp data.
		out *byte = nil
	)
	size := lib.Encode(
		libc.NewTLS(),
		uintptr(unsafe.Pointer(&rgbaImage.Pix[0])),
		int32(rect.Dx()),
		int32(rect.Dy()),
		int32(rgbaImage.Stride),
		// Function pointers are generated by taking a pointer to
		// a struct who's first field is that function.
		*(*uintptr)(unsafe.Pointer(
			&struct {
				f func(tls *libc.TLS, picture uintptr, rgba uintptr, rgba_stride int32) int32
			}{
				lib.WebPPictureImportRGBA,
			},
		)),
		// quality for libwebp is [0,100].
		float32(enc.Quality*100),
		boolToInt32(enc.Lossless),
		uintptr(unsafe.Pointer(&out)),
	)
	if size == 0 {
		return fmt.Errorf("encoding webp image: size %d", size)
	}
	if out == nil {
		return fmt.Errorf("failed to allocate memory; probably errored")
	}
	if _, err := w.Write(libc.GoBytes(uintptr(unsafe.Pointer(out)), int(size))); err != nil {
		return fmt.Errorf("writing webp data: %w", err)
	}
	return nil
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}