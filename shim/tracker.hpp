#pragma once
#include <cuda.h>
#include <cstdint>
#include <functional>
#include <mutex>
#include <string>
#include <unordered_map>
#include <vector>

struct AllocRec {
  CUdeviceptr base;
  size_t size;
  CUcontext ctx;
  unsigned flags;
  uint64_t seq;
};

struct StreamRec {
  CUstream stream;
  CUcontext ctx;
  unsigned flags;
};

struct CtxRec {
  CUcontext ctx;
  CUdevice dev;
  unsigned flags;
};

class Tracker {
 public:
  static Tracker& instance();
  void track_alloc(CUdeviceptr p, size_t sz, CUcontext ctx, unsigned fl);
  void untrack_alloc(CUdeviceptr p);
  bool lookup_alloc(CUdeviceptr p, AllocRec* out) const;
  std::vector<AllocRec> allocs_snapshot() const;
  size_t alloc_count() const;
  uint64_t total_bytes() const;
  void track_stream(CUstream s, CUcontext ctx, unsigned fl);
  void untrack_stream(CUstream s);
  std::vector<StreamRec> streams_snapshot() const;
  void track_ctx(CUcontext c, CUdevice d, unsigned fl);
  void untrack_ctx(CUcontext c);
  std::vector<CtxRec> ctxs_snapshot() const;
  void set_primary(CUcontext c);
  CUcontext primary() const;
  void set_device(CUdevice d);
  CUdevice device() const;
  bool unsupported_detected() const;
  void mark_unsupported(const char* reason);
  std::string unsupported_reason() const;
  void clear();
  void set_image_dir(const std::string& d);
  std::string image_dir() const;
  void set_state(int s);
  int state() const;

 private:
  Tracker() = default;
  static constexpr int kShards = 64;
  struct Shard {
    mutable std::mutex mu;
    std::unordered_map<CUdeviceptr, AllocRec> map;
  };
  Shard shards_[kShards];
  mutable std::mutex stream_mu_;
  std::unordered_map<CUstream, StreamRec> streams_;
  mutable std::mutex ctx_mu_;
  std::unordered_map<CUcontext, CtxRec> ctxs_;
  CUcontext primary_{nullptr};
  CUdevice device_{0};
  mutable std::mutex meta_mu_;
  bool bad_{false};
  std::string bad_reason_;
  std::string image_dir_;
  int state_{0};
  uint64_t next_seq_{1};
  Shard& shard(CUdeviceptr p);
  const Shard& shard(CUdeviceptr p) const;
};
