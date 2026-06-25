#pragma once
#include <cuda.h>
#include <cstdio>
#include <cstdlib>

#define CUDA_CHECK(call)                                                     \
  do {                                                                       \
    CUresult _r = (call);                                                    \
    if (_r != CUDA_SUCCESS) {                                                \
      const char* _e = nullptr;                                              \
      cuGetErrorString(_r, &_e);                                             \
      fprintf(stderr, "cudackpt: %s:%d %s\n", __FILE__, __LINE__,            \
              _e ? _e : "cuda error");                                       \
      abort();                                                               \
    }                                                                        \
  } while (0)
