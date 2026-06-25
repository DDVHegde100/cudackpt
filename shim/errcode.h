#pragma once
#include <stdint.h>

typedef enum {
  CKPT_ERR_NONE = 0,
  CKPT_ERR_NCCL = 100,
  CKPT_ERR_CUDA_GRAPH = 101,
  CKPT_ERR_MIG = 102,
  CKPT_ERR_MULTI_GPU = 103,
  CKPT_ERR_PEER_ACCESS = 104,
  CKPT_ERR_GENERIC = 199
} CkptErrCode;

const char* ckpt_err_name(uint32_t code);
