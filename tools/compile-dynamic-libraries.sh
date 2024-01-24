#!/bin/sh

VERSION="7.1.3"

BASE="./c-lib"
BUILD=$BASE/build
BUILD_LINUX=$BUILD-linux
BUILD_WINDOWS=$BUILD-windows
BUILD_MACOS=$BUILD-macos

set -eux

function prepare() {
  rm -r $1 || true
  mkdir -p $1
}

# Linux

prepare $BUILD_LINUX

( 
  cd $BUILD_LINUX                                        
  cmake -DBUILD_SHARED_LIBS="ON" -DCMAKE_TOOLCHAIN_FILE=../../tools/zig-toolchain-x86_64.linux.cmake ../ 
  make webp                                              
  cp libwebp.so.$VERSION ../../lib/dynamic/webp/blobs/libwebp.so 
)

# Windows

prepare $BUILD_WINDOWS

( 
  cd $BUILD_WINDOWS                                                                                           
  cmake -DBUILD_SHARED_LIBS="ON" -DCMAKE_TOOLCHAIN_FILE=../../tools/zig-toolchain-x86_64.windows.cmake ../ 
  make webp                                                                                                   
  cp libwebp.dll ../../lib/dynamic/webp/blobs/libwebp.dll                                                          
)

# macOS

prepare $BUILD_MACOS

(
  cd $BUILD_MACOS                                                                                         
  cmake -DBUILD_SHARED_LIBS="ON" -DCMAKE_TOOLCHAIN_FILE=../../tools/zig-toolchain-x86_64.macos.cmake ../ 
  make webp                                                                                               
  cp libwebp.$VERSION.dylib ../../lib/dynamic/webp/blobs/libwebp.dylib                                         
)

