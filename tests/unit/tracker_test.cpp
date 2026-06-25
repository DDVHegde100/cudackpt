#include "tracker.hpp"
#include <cassert>
#include <thread>
#include <vector>

int main() {
  auto& tr = Tracker::instance();
  tr.clear();
  tr.track_alloc(0x1000, 4096, reinterpret_cast<CUcontext>(1), 0);
  tr.track_alloc(0x2000, 8192, reinterpret_cast<CUcontext>(1), 0);
  AllocRec rec{};
  assert(tr.lookup_alloc(0x1000, &rec));
  assert(rec.size == 4096);
  tr.untrack_alloc(0x1000);
  assert(!tr.lookup_alloc(0x1000, nullptr));
  std::vector<std::thread> ts;
  for (int i = 0; i < 8; ++i) {
    ts.emplace_back([i, &tr]() {
      CUdeviceptr p = static_cast<CUdeviceptr>(0x10000 + i * 0x1000);
      tr.track_alloc(p, 1024, reinterpret_cast<CUcontext>(1), 0);
      assert(tr.lookup_alloc(p, nullptr));
      tr.untrack_alloc(p);
    });
  }
  for (auto& t : ts) t.join();
  tr.track_stream(reinterpret_cast<CUstream>(0x10), reinterpret_cast<CUcontext>(1), 0);
  assert(tr.streams_snapshot().size() == 1);
  tr.track_ctx(reinterpret_cast<CUcontext>(2), 0, 0);
  assert(tr.ctxs_snapshot().size() == 1);
  return 0;
}
