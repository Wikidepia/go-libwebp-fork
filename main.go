package main

import (
	"encoding/binary"
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
		quality = float32(1.0)
		// BUG: turning this off leads to divide by zero.
		lossless = int32(1)
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

// bsearch stub that ostensibly implements a binary search on unsafe data.
//
// NOTE(jfm): current implementation is actually just a linear search
// that ignores the predicate function for now.
//
// Input data is expected to be sorted low to high, and contain i32.
//
// Assuming i32 is technically eroneous and we should use the predicate
// function to capture type information to keep this function data agnostic.
// However it might be better to simply implement a Go idiomatic function
// that doesn't operate on raw memory.
//
// Given that this function is used exactly once and the data used is i32
// this works.
//
//       key: raw pointer to object that serves as key
//      base: raw pointer to first element in array
//  elements: number of elements in array
//      size: size of each element in bytes
// predicate: comparison function applied to each set of candidates [type predicate func(left, right uintptr) int32]
//
func bsearch(tls *libc.TLS, key, base uintptr, elements, size uint64, predicate uintptr) uintptr {
	var (
		// This conversion from uintptr to unsafe.Pointer may actually be valid based
		// on the codebase using uintptr exclusively, the memory is not GC'd like standard Go
		// but rather allocated with `libc.Alloc`.
		k    int32 = *(*int32)(unsafe.Pointer(key))
		data       = libc.GoBytes(base, int(elements*size))
	)
	// Linear search for exact match.
	for ii := uint64(0); ii < uint64(len(data)); ii += size {
		n := int32(binary.LittleEndian.Uint32(data[ii : ii+size]))
		if n == k {
			// return the base pointer offset by the number of bytes
			// at which the value exists.
			return base + uintptr(ii)
		}
	}
	// return the base memory offset if we didn't find a match. May not
	// be correct semantically, but should avoid memory corruption.
	return base
}
