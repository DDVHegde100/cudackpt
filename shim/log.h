#pragma once
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>

static inline void ckpt_log(const char* fmt, ...) {
  if (!getenv("CUDACKPT_DEBUG")) return;
  va_list ap;
  va_start(ap, fmt);
  vfprintf(stderr, fmt, ap);
  va_end(ap);
}
