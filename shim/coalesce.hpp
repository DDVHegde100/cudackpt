#pragma once
#include "tracker.hpp"
#include <vector>

struct CoalescedAlloc {
  CUdeviceptr base;
  size_t size;
  CUcontext ctx;
  unsigned flags;
  uint64_t seq;
  std::vector<AllocRec> parts;
};

std::vector<CoalescedAlloc> coalesce_allocs(const std::vector<AllocRec>& allocs);
