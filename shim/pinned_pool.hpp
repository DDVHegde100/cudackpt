#pragma once
#include <cuda.h>
#include <cstddef>
#include <mutex>
#include <vector>

class PinnedPool {
 public:
  static PinnedPool& instance();
  void* acquire(size_t nbytes);
  void release(void* ptr);
  void clear();

 private:
  struct Block {
    void* ptr;
    size_t cap;
    bool free;
  };
  std::mutex mu_;
  std::vector<Block> blocks_;
};
