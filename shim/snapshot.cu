#include "ckpt_ops.h"
#include "log.h"
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
  AllocRec alloc;
  uint64_t offset;
  std::vector<uint8_t> host;
  uint32_t crc;
  int err_code;
  CUresult err;
};

static void snap_worker(const std::vector<AllocRec>* allocs, std::vector<SnapTask>* tasks,
                        std::atomic<size_t>* next, std::atomic<int>* fail) {
  for (;;) {
    size_t i = next->fetch_add(1);
    if (i >= allocs->size() || fail->load() != 0) return;
    SnapTask& t = (*tasks)[i];
    t.alloc = (*allocs)[i];
    t.host.resize(t.alloc.size);
    CUcontext prev = nullptr;
    CUresult r = cuCtxGetCurrent(&prev);
    if (r != CUDA_SUCCESS) {
      t.err_code = -13;
      t.err = r;
      fail->store(t.err_code);
      return;
    }
    if (t.alloc.ctx) {
      r = cuCtxSetCurrent(t.alloc.ctx);
      if (r != CUDA_SUCCESS) {
        t.err_code = -14;
        t.err = r;
        fail->store(t.err_code);
        return;
      }
    }
    r = cuMemcpyDtoH(t.host.data(), t.alloc.base, t.alloc.size);
    if (prev) cuCtxSetCurrent(prev);
    if (r != CUDA_SUCCESS) {
      t.err_code = -15;
      t.err = r;
      fail->store(t.err_code);
      return;
    }
    t.crc = crc32c(t.host.data(), t.host.size());
  }
}

int ckpt_snapshot_write(const char* dir, ChunkEntry* entries, int* count) {
  auto& tr = Tracker::instance();
  if (tr.unsupported_detected()) return snapshot_fail(dir, -1, CUDA_SUCCESS, "unsupported");
  auto streams = tr.streams_snapshot();
  for (const auto& sr : streams) {
    CUresult sr_r = cuStreamSynchronize(sr.stream);
    if (sr_r != CUDA_SUCCESS) return snapshot_fail(dir, -12, sr_r, "stream sync");
  }
  auto allocs = tr.allocs_snapshot();
  std::vector<SnapTask> tasks(allocs.size());
  uint64_t off = 0;
  for (size_t i = 0; i < allocs.size(); ++i) {
    tasks[i].offset = off;
    off += allocs[i].size;
  }
  unsigned hw = std::thread::hardware_concurrency();
  unsigned workers = hw == 0 ? 4u : std::min(8u, hw);
  if (workers > static_cast<unsigned>(allocs.size())) workers = static_cast<unsigned>(allocs.size());
  if (workers == 0) workers = 1;
  std::atomic<size_t> next{0};
  std::atomic<int> fail{0};
  std::vector<std::thread> pool;
  pool.reserve(workers);
  for (unsigned w = 0; w < workers; ++w)
    pool.emplace_back(snap_worker, &allocs, &tasks, &next, &fail);
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
    out.write(reinterpret_cast<const char*>(t.host.data()),
              static_cast<std::streamsize>(t.host.size()));
    manifest.push_back(ChunkEntry{static_cast<uint64_t>(t.alloc.base),
                                  static_cast<uint64_t>(t.alloc.size), t.offset, t.crc,
                                  static_cast<uint32_t>(t.alloc.seq)});
  }
  std::ofstream mf(path_join(dir, "manifest.bin"), std::ios::binary | std::ios::trunc);
  if (!mf) return snapshot_fail(dir, -3, CUDA_SUCCESS, "manifest.bin");
  struct {
    uint32_t magic;
    uint16_t version;
    uint16_t flags;
    uint32_t count;
    uint64_t total;
  } hdr{0x434B5054u, 1, 0, static_cast<uint32_t>(manifest.size()), off};
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
  unlink(path_join(dir, "snapshot.err").c_str());
  tr.set_state(2);
  ckpt_log("cudackpt: snapshot ok count=%zu bytes=%llu workers=%u\n", manifest.size(),
           (unsigned long long)off, workers);
  return 0;
}
