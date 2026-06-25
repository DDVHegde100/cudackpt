#include "pinned_pool.hpp"
#include <algorithm>

PinnedPool& PinnedPool::instance() {
  static PinnedPool p;
  return p;
}

void* PinnedPool::acquire(size_t nbytes) {
  if (nbytes == 0) return nullptr;
  std::lock_guard<std::mutex> lk(mu_);
  for (auto& b : blocks_) {
    if (b.free && b.cap >= nbytes) {
      b.free = false;
      return b.ptr;
    }
  }
  void* ptr = nullptr;
  if (cuMemHostAlloc(&ptr, nbytes, CU_MEMHOSTALLOC_PORTABLE) != CUDA_SUCCESS) return nullptr;
  blocks_.push_back(Block{ptr, nbytes, false});
  return ptr;
}

void PinnedPool::release(void* ptr) {
  if (!ptr) return;
  std::lock_guard<std::mutex> lk(mu_);
  for (auto& b : blocks_) {
    if (b.ptr == ptr) {
      b.free = true;
      return;
    }
  }
}

void PinnedPool::clear() {
  std::lock_guard<std::mutex> lk(mu_);
  for (auto& b : blocks_) {
    if (b.ptr) cuMemFreeHost(b.ptr);
  }
  blocks_.clear();
}
