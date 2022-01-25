package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"unsafe"

	_ "image/jpeg"
	_ "image/png"

	"modernc.org/libc"
)

func main() {
	runtime.LockOSThread()
	if err := run(os.Args[1:]); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("provide image to encode as webp, try megopher.png")
	}
	input := args[0]
	srcf, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("opening input file: %w", err)
	}
	defer srcf.Close()
	img, _, err := image.Decode(srcf)
	if err != nil {
		return fmt.Errorf("decoding src image: %w", err)
	}
	rgbaImg := image.NewNRGBA(img.Bounds())
	rect := img.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}
	var (
		quality  = float32(100.0)
		lossless = int32(0)
		// out buffer to contain webp data.
		out *byte = nil
	)
	size := Encode(
		libc.NewTLS(),
		uintptr(unsafe.Pointer(&rgbaImg.Pix[0])),
		int32(img.Bounds().Dx()),
		int32(img.Bounds().Dy()),
		int32(rgbaImg.Stride),
		// Function pointers are generated by taking a pointer to
		// a struct who's first field is that function.
		// That being said, this could be nil because we don't actually
		// invoke it in our bsearch implementation.
		*(*uintptr)(unsafe.Pointer(
			&struct {
				f func(tls *libc.TLS, picture uintptr, rgba uintptr, rgba_stride int32) int32
			}{
				WebPPictureImportRGBA,
			},
		)),
		quality,
		lossless,
		uintptr(unsafe.Pointer(&out)),
	)
	if size == 0 {
		return fmt.Errorf("encoding webp image: size %d", size)
	}
	if out == nil {
		return fmt.Errorf("failed to allocate memory; probably errored")
	}
	if err := os.WriteFile("out.webp", libc.GoBytes(uintptr(unsafe.Pointer(out)), int(size)), 0644); err != nil {
		return fmt.Errorf("writing out webp image: %w", err)
	}
	return nil
}
