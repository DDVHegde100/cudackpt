#include "errcode.h"

const char* ckpt_err_name(uint32_t code) {
  switch (code) {
    case CKPT_ERR_NONE:
      return "none";
    case CKPT_ERR_NCCL:
      return "nccl";
    case CKPT_ERR_CUDA_GRAPH:
      return "cuda_graph";
    case CKPT_ERR_MIG:
      return "mig";
    case CKPT_ERR_MULTI_GPU:
      return "multi_gpu";
    case CKPT_ERR_PEER_ACCESS:
      return "peer_access";
    default:
      return "unsupported";
  }
}
