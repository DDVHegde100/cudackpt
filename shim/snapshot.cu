#include "ckpt_ops.h"
#include "coalesce.hpp"
#include "log.h"
#include "module_io.hpp"
#include "pinned_pool.hpp"
#include "quiesce.hpp"
#include "tracker.hpp"
#include <algorithm>
#include <atomic>
#include <cstdio>
#include <cstring>
#include <fstream>
#include <string>
#include <thread>
#include <unistd.h>
#include <vector>

static uint32_t crc32c_table[256];
static bool crc_init = false;

static void init_crc() {
  if (crc_init) return;
  for (uint32_t i = 0; i < 256; ++i) {
    uint32_t c = i;
    for (int k = 0; k < 8; ++k)
      c = (c & 1) ? (0x82F63B78u ^ (c >> 1)) : (c >> 1);
    crc32c_table[i] = c;
  }
  crc_init = true;
}

static uint32_t crc32c(const void* data, size_t len) {
  init_crc();
  const auto* p = static_cast<const uint8_t*>(data);
  uint32_t c = 0xFFFFFFFFu;
  for (size_t i = 0; i < len; ++i)
    c = crc32c_table[(c ^ p[i]) & 0xFFu] ^ (c >> 8);
  return c ^ 0xFFFFFFFFu;
}

static std::string path_join(const char* dir, const char* name) {
  std::string s(dir);
  if (!s.empty() && s.back() != '/') s += '/';
  s += name;
  return s;
}

static int snapshot_fail(const char* dir, int code, CUresult r, const char* msg) {
  std::ofstream ef(path_join(dir, "snapshot.err"), std::ios::binary | std::ios::trunc);
  const char* es = nullptr;
  if (r != CUDA_SUCCESS) cuGetErrorString(r, &es);
  if (ef) {
    ef << code;
    if (msg) ef << " " << msg;
    if (es) ef << " " << es;
    ef << "\n";
  }
  ckpt_log("cudackpt: snapshot fail %d %s %s\n", code, msg ? msg : "", es ? es : "");
  return code;
}

struct SnapTask {
  CoalescedAlloc block;
  uint64_t offset;
  void* host;
  size_t host_cap;
  uint32_t crc;
  int err_code;
  CUresult err;
};

static void snap_worker(const std::vector<CoalescedAlloc>* blocks, std::vector<SnapTask>* tasks,
                        std::atomic<size_t>* next, std::atomic<int>* fail) {
  auto& pool = PinnedPool::instance();
  for (;;) {
    size_t i = next->fetch_add(1);
    if (i >= blocks->size() || fail->load() != 0) return;
    SnapTask& t = (*tasks)[i];
    t.block = (*blocks)[i];
    t.host_cap = t.block.size;
    t.host = pool.acquire(t.host_cap);
    if (!t.host) {
      t.err_code = -16;
      t.err = CUDA_ERROR_OUT_OF_MEMORY;
      fail->store(t.err_code);
      return;
    }
    CUcontext prev = nullptr;
    CUresult r = cuCtxGetCurrent(&prev);
    if (r != CUDA_SUCCESS) {
      t.err_code = -13;
      t.err = r;
      fail->store(t.err_code);
      return;
    }
    if (t.block.ctx) {
      r = cuCtxSetCurrent(t.block.ctx);
      if (r != CUDA_SUCCESS) {
        t.err_code = -14;
        t.err = r;
        fail->store(t.err_code);
        return;
      }
    }
    r = cuMemcpyDtoH(t.host, t.block.base, t.block.size);
    if (prev) cuCtxSetCurrent(prev);
    if (r != CUDA_SUCCESS) {
      t.err_code = -15;
      t.err = r;
      fail->store(t.err_code);
      return;
    }
    t.crc = crc32c(t.host, t.block.size);
  }
}

int ckpt_snapshot_write(const char* dir, ChunkEntry* entries, int* count) {
  auto& tr = Tracker::instance();
  if (tr.unsupported_detected()) {
    ckpt_log("cudackpt: unsupported code=%u reason=%s\n", tr.unsupported_code(),
             tr.unsupported_reason().c_str());
    return snapshot_fail(dir, -static_cast<int>(tr.unsupported_code()), CUDA_SUCCESS, "unsupported");
  }
  if (ckpt_quiesce_gpu() != 0) return snapshot_fail(dir, -25, CUDA_SUCCESS, "quiesce");
  auto allocs = tr.allocs_snapshot();
  auto blocks = coalesce_allocs(allocs);
  std::vector<SnapTask> tasks(blocks.size());
  uint64_t off = 0;
  for (size_t i = 0; i < blocks.size(); ++i) {
    tasks[i].offset = off;
    off += blocks[i].size;
  }
  unsigned hw = std::thread::hardware_concurrency();
  unsigned workers = hw == 0 ? 4u : std::min(8u, hw);
  if (workers > static_cast<unsigned>(blocks.size())) workers = static_cast<unsigned>(blocks.size());
  if (workers == 0) workers = 1;
  std::atomic<size_t> next{0};
  std::atomic<int> fail{0};
  std::vector<std::thread> pool;
  pool.reserve(workers);
  for (unsigned w = 0; w < workers; ++w)
    pool.emplace_back(snap_worker, &blocks, &tasks, &next, &fail);
  for (auto& th : pool) th.join();
  if (fail.load() != 0) {
    for (const auto& t : tasks) {
      if (t.err_code != 0) return snapshot_fail(dir, t.err_code, t.err, "dtoh");
    }
  }
  std::ofstream out(path_join(dir, "device.bin"), std::ios::binary | std::ios::trunc);
  if (!out) return snapshot_fail(dir, -2, CUDA_SUCCESS, "device.bin");
  std::vector<ChunkEntry> manifest;
  manifest.reserve(tasks.size());
  for (const auto& t : tasks) {
    out.write(static_cast<const char*>(t.host), static_cast<std::streamsize>(t.block.size));
    for (const auto& part : t.block.parts) {
      uint64_t rel = static_cast<uint64_t>(part.base) - static_cast<uint64_t>(t.block.base);
      uint32_t pcrc = crc32c(static_cast<uint8_t*>(t.host) + rel, part.size);
      manifest.push_back(ChunkEntry{static_cast<uint64_t>(part.base), static_cast<uint64_t>(part.size),
                                    t.offset + rel, pcrc, static_cast<uint32_t>(part.seq)});
    }
    if (t.host) PinnedPool::instance().release(t.host);
  }
  std::ofstream mf(path_join(dir, "manifest.bin"), std::ios::binary | std::ios::trunc);
  if (!mf) return snapshot_fail(dir, -3, CUDA_SUCCESS, "manifest.bin");
  struct {
    uint32_t magic;
    uint16_t version;
    uint16_t flags;
    uint32_t count;
    uint64_t total;
  } hdr{0x434B5054u, 2, 0, static_cast<uint32_t>(manifest.size()), off};
  mf.write(reinterpret_cast<const char*>(&hdr), sizeof(hdr));
  mf.write(reinterpret_cast<const char*>(manifest.data()),
           static_cast<std::streamsize>(manifest.size() * sizeof(ChunkEntry)));
  if (entries && count) {
    int n = static_cast<int>(manifest.size());
    for (int i = 0; i < n; ++i) entries[i] = manifest[static_cast<size_t>(i)];
    *count = n;
  } else if (count) {
    *count = static_cast<int>(manifest.size());
  }
  {
    uint32_t dev = static_cast<uint32_t>(tr.device());
    std::ofstream df(path_join(dir, "dev.bin"), std::ios::binary | std::ios::trunc);
    if (!df) return snapshot_fail(dir, -4, CUDA_SUCCESS, "dev.bin");
    df.write(reinterpret_cast<const char*>(&dev), sizeof(dev));
  }
  if (ckpt_ctx_write(dir) != 0) return snapshot_fail(dir, -41, CUDA_SUCCESS, "ctx.bin");
  if (ckpt_streams_write(dir) != 0) return snapshot_fail(dir, -51, CUDA_SUCCESS, "streams.bin");
  if (ckpt_modules_write(dir) != 0) return snapshot_fail(dir, -34, CUDA_SUCCESS, "modules.bin");
  unlink(path_join(dir, "snapshot.err").c_str());
  tr.set_state(2);
  ckpt_log("cudackpt: snapshot ok chunks=%zu manifest=%zu bytes=%llu workers=%u\n", blocks.size(),
           manifest.size(), (unsigned long long)off, workers);
  return 0;
}
