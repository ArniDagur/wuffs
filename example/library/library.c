// Copyright 2017 The Wuffs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
library exercises the software libraries built by `wuffs genlib`.

To exercise the static library:

$CC -static -I../../.. library.c ../../gen/lib/c/$CC-static/libwuffs.a
./a.out
rm -f a.out

To exercise the dynamic library:

$CC -I../../.. library.c -L../../gen/lib/c/$CC-dynamic -lwuffs
LD_LIBRARY_PATH=../../gen/lib/c/$CC-dynamic ./a.out
rm -f a.out

for a C compiler $CC, such as clang or gcc.
*/

#include <stdlib.h>
#include <unistd.h>

#include "wuffs/release/c/wuffs-unsupported-snapshot.h"

#define DST_BUFFER_SIZE (1024 * 1024)

// lgtm_ptr and lgtm_len hold a deflate-encoded "LGTM" message.
uint8_t lgtm_ptr[] = {
    0xf3, 0xc9, 0xcf, 0xcf, 0x2e, 0x56, 0x48, 0xcf, 0xcf, 0x4f,
    0x51, 0x28, 0xc9, 0x57, 0xc8, 0x4d, 0xd5, 0xe3, 0x02, 0x00,
};
size_t lgtm_len = 20;

// ignore_return_value suppresses errors from -Wall -Werror.
static void ignore_return_value(int ignored) {}

static const char* decode() {
  uint8_t dst_buffer[DST_BUFFER_SIZE];
  wuffs_base__io_buffer dst = ((wuffs_base__io_buffer){
      .data = ((wuffs_base__slice_u8){
          .ptr = dst_buffer,
          .len = DST_BUFFER_SIZE,
      }),
  });
  wuffs_base__io_buffer src = ((wuffs_base__io_buffer){
      .data = ((wuffs_base__slice_u8){
          .ptr = lgtm_ptr,
          .len = lgtm_len,
      }),
      .meta = ((wuffs_base__io_buffer_meta){
          .wi = lgtm_len,
          .ri = 0,
          .pos = 0,
          .closed = true,
      }),
  });
  wuffs_base__io_writer dst_writer = wuffs_base__io_buffer__writer(&dst);
  wuffs_base__io_reader src_reader = wuffs_base__io_buffer__reader(&src);

  wuffs_deflate__decoder* dec = calloc(sizeof__wuffs_deflate__decoder(), 1);
  if (!dec) {
    return "out of memory";
  }
  wuffs_base__status z = wuffs_deflate__decoder__check_wuffs_version(
      dec, sizeof__wuffs_deflate__decoder(), WUFFS_VERSION);
  if (z) {
    free(dec);
    return z;
  }
  z = wuffs_deflate__decoder__decode(dec, dst_writer, src_reader);
  if (z) {
    free(dec);
    return z;
  }
  ignore_return_value(write(1, dst.data.ptr, dst.meta.wi));
  free(dec);
  return NULL;
}

int main(int argc, char** argv) {
  const char* status_msg = decode();
  int status = 0;
  if (status_msg) {
    status = 1;
    ignore_return_value(write(2, status_msg, strnlen(status_msg, 4095)));
    ignore_return_value(write(2, "\n", 1));
  }
  return status;
}
