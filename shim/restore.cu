#include "ckpt_ops.h"
#include "log.h"
#include "module_io.hpp"
#include "pinned_pool.hpp"
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

static std::string path_join(const char* dir, const char* name) {
  std::string s(dir);
  if (!s.empty() && s.back() != '/') s += '/';
  s += name;
  return s;
}

static int restore_fail(const char* dir, int code, CUresult r, const char* msg) {
  std::ofstream ef(path_join(dir, "restore.err"), std::ios::binary | std::ios::trunc);
  const char* es = nullptr;
  if (r != CUDA_SUCCESS) cuGetErrorString(r, &es);
  if (ef) {
    ef << code;
    if (msg) ef << " " << msg;
    if (es) ef << " " << es;
    ef << "\n";
  }
  ckpt_log("cudackpt: restore fail %d %s %s\n", code, msg ? msg : "", es ? es : "");
  return code;
}

static CUresult map_at(CUdevice dev, CUdeviceptr want, size_t sz) {
  CUmemAllocationProp prop{};
  prop.type = CU_MEM_ALLOCATION_TYPE_PINNED;
  prop.location.type = CU_MEM_LOCATION_TYPE_DEVICE;
  prop.location.id = dev;
  size_t gran = 0;
  CUresult gr = cuMemGetAllocationGranularity(&gran, &prop, CU_MEM_ALLOC_GRANULARITY_MINIMUM);
  if (gr != CUDA_SUCCESS) return gr;
  size_t aligned = ((sz + gran - 1) / gran) * gran;
  CUdeviceptr addr = want;
  CUresult r = cuMemAddressReserve(&addr, aligned, gran, want, 0);
  if (r != CUDA_SUCCESS) return r;
  CUmemGenericAllocationHandle h;
  r = cuMemCreate(&h, aligned, &prop, 0);
  if (r != CUDA_SUCCESS) {
    cuMemAddressFree(addr, aligned);
    return r;
  }
  r = cuMemMap(addr, aligned, 0, h, 0);
  if (r != CUDA_SUCCESS) {
    cuMemRelease(h);
    cuMemAddressFree(addr, aligned);
    return r;
  }
  CUmemAccessDesc acc{};
  acc.location.type = CU_MEM_LOCATION_TYPE_DEVICE;
  acc.location.id = dev;
  acc.flags = CU_MEM_ACCESS_FLAGS_PROT_READWRITE;
  r = cuMemSetAccess(addr, aligned, &acc, 1);
  if (r != CUDA_SUCCESS) {
    cuMemUnmap(addr, aligned);
    cuMemRelease(h);
    cuMemAddressFree(addr, aligned);
    return r;
  }
  if (addr != want) {
    cuMemUnmap(addr, aligned);
    cuMemRelease(h);
    cuMemAddressFree(addr, aligned);
    return CUDA_ERROR_INVALID_VALUE;
  }
  return CUDA_SUCCESS;
}

struct RestoreTask {
  ChunkEntry entry;
  CUdeviceptr ptr;
  void* host;
  size_t host_cap;
  int err_code;
  CUresult err;
};

static void restore_htod_worker(std::vector<RestoreTask>* tasks, std::atomic<size_t>* next,
                                std::atomic<int>* fail) {
  for (;;) {
    size_t i = next->fetch_add(1);
    if (i >= tasks->size() || fail->load() != 0) return;
    RestoreTask& t = (*tasks)[i];
    CUresult r = cuMemcpyHtoD(t.ptr, t.host, static_cast<size_t>(t.entry.size));
    if (r != CUDA_SUCCESS) {
      t.err_code = -10;
      t.err = r;
      fail->store(t.err_code);
      return;
    }
  }
}

int ckpt_restore_load(const char* dir, ChunkEntry* entries, int max, int* count) {
  auto& tr = Tracker::instance();
  std::string manifest = path_join(dir, "manifest.bin");
  std::ifstream mf(manifest, std::ios::binary);
  if (!mf) return restore_fail(dir, -1, CUDA_SUCCESS, "manifest");
  struct {
    uint32_t magic;
    uint16_t version;
    uint16_t flags;
    uint32_t count;
    uint64_t total;
  } hdr{};
  mf.read(reinterpret_cast<char*>(&hdr), sizeof(hdr));
  if (hdr.magic != 0x434B5054u) return restore_fail(dir, -2, CUDA_SUCCESS, "magic");
  std::vector<ChunkEntry> want(hdr.count);
  mf.read(reinterpret_cast<char*>(want.data()),
          static_cast<std::streamsize>(hdr.count * sizeof(ChunkEntry)));
  if (!mf) return restore_fail(dir, -3, CUDA_SUCCESS, "manifest body");
  bool have_seq = false;
  for (const auto& e : want) {
    if (e.pad != 0) have_seq = true;
  }
  if (have_seq) {
    std::stable_sort(want.begin(), want.end(),
                     [](const ChunkEntry& a, const ChunkEntry& b) { return a.pad < b.pad; });
  }
  std::ifstream df(path_join(dir, "device.bin"), std::ios::binary);
  if (!df) return restore_fail(dir, -4, CUDA_SUCCESS, "device.bin");
  CUdevice dev = tr.device();
  if (dev == 0) {
    std::ifstream devf(path_join(dir, "dev.bin"), std::ios::binary);
    if (devf) {
      uint32_t d = 0;
      devf.read(reinterpret_cast<char*>(&d), sizeof(d));
      dev = static_cast<CUdevice>(d);
      tr.set_device(dev);
    }
  }
  CUresult r = cuInit(0);
  if (r != CUDA_SUCCESS) return restore_fail(dir, -6, r, "cuInit");
  ckpt_ctx_restore(dir);
  CUcontext ctx = tr.primary();
  if (!ctx) {
    if (cuDevicePrimaryCtxRetain(&ctx, dev) != CUDA_SUCCESS)
      return restore_fail(dir, -7, r, "ctx retain");
  }
  r = cuCtxSetCurrent(ctx);
  if (r != CUDA_SUCCESS) return restore_fail(dir, -8, r, "ctx set");
  tr.set_primary(ctx);
  tr.set_device(dev);
  tr.clear();
  tr.track_ctx(ctx, dev, tr.primary_flags());
  if (ckpt_modules_restore(dir) != 0) return restore_fail(dir, -35, CUDA_SUCCESS, "modules");
  ckpt_streams_restore(dir);
  auto& pool = PinnedPool::instance();
  std::vector<RestoreTask> tasks(want.size());
  for (size_t i = 0; i < want.size(); ++i) {
    const auto& e = want[i];
    tasks[i].entry = e;
    CUdeviceptr p = 0;
    CUresult ar = cuMemAlloc(&p, static_cast<size_t>(e.size));
    if (ar != CUDA_SUCCESS || static_cast<uint64_t>(p) != e.ptr) {
      if (p) cuMemFree(p);
      CUresult mr = map_at(dev, static_cast<CUdeviceptr>(e.ptr), static_cast<size_t>(e.size));
      if (mr != CUDA_SUCCESS)
        return restore_fail(dir, -5, mr, "addr remap");
      p = static_cast<CUdeviceptr>(e.ptr);
    }
    tasks[i].ptr = p;
    tr.track_alloc(p, static_cast<size_t>(e.size), ctx, 0);
    tasks[i].host_cap = static_cast<size_t>(e.size);
    tasks[i].host = pool.acquire(tasks[i].host_cap);
    if (!tasks[i].host) return restore_fail(dir, -17, CUDA_ERROR_OUT_OF_MEMORY, "pinned");
    df.seekg(static_cast<std::streamoff>(e.offset));
    df.read(static_cast<char*>(tasks[i].host), static_cast<std::streamsize>(e.size));
    if (!df) return restore_fail(dir, -9, CUDA_SUCCESS, "device read");
  }
  unsigned hw = std::thread::hardware_concurrency();
  unsigned workers = hw == 0 ? 4u : std::min(8u, hw);
  if (workers > static_cast<unsigned>(tasks.size())) workers = static_cast<unsigned>(tasks.size());
  if (workers == 0) workers = 1;
  std::atomic<size_t> next{0};
  std::atomic<int> fail{0};
  std::vector<std::thread> thpool;
  thpool.reserve(workers);
  for (unsigned w = 0; w < workers; ++w)
    thpool.emplace_back(restore_htod_worker, &tasks, &next, &fail);
  for (auto& th : thpool) th.join();
  if (fail.load() != 0) {
    for (const auto& t : tasks) {
      if (t.err_code != 0) return restore_fail(dir, t.err_code, t.err, "htod");
    }
  }
  for (auto& t : tasks) {
    if (t.host) pool.release(t.host);
  }
  r = cuCtxSynchronize();
  if (r != CUDA_SUCCESS) return restore_fail(dir, -11, r, "sync");
  if (entries && max > 0) {
    int n = std::min(max, static_cast<int>(want.size()));
    std::memcpy(entries, want.data(), static_cast<size_t>(n) * sizeof(ChunkEntry));
    if (count) *count = n;
  } else if (count) {
    *count = static_cast<int>(want.size());
  }
  tr.set_state(3);
  unlink(path_join(dir, "restore.err").c_str());
  ckpt_log("cudackpt: restore ok count=%u workers=%u\n", hdr.count, workers);
  return 0;
}

int ckpt_resume_signal() {
  Tracker::instance().set_state(4);
  return 0;
}
