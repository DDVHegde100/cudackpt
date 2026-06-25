#pragma once
#include <cuda.h>
#include <cstdint>
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
  int priority;
  unsigned capture_status;
};

struct CtxRec {
  CUcontext ctx;
  CUdevice dev;
  unsigned flags;
};

struct ModuleRec {
  CUmodule mod;
  CUcontext ctx;
  std::string path;
  std::vector<uint8_t> image;
  uint64_t seq;
};

struct SymbolRec {
  CUfunction fn;
  CUmodule mod;
  std::string name;
};

struct EventRec {
  CUevent event;
  CUcontext ctx;
  unsigned flags;
};

struct TrackerStats {
  size_t alloc_count;
  uint64_t total_bytes;
  size_t stream_count;
  size_t module_count;
  size_t symbol_count;
  size_t event_count;
  size_t ctx_count;
  uint32_t unsupported_code;
  int state;
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
  void update_stream_state(CUstream s, int priority, unsigned capture_status);
  std::vector<StreamRec> streams_snapshot() const;
  void track_ctx(CUcontext c, CUdevice d, unsigned fl);
  void untrack_ctx(CUcontext c);
  std::vector<CtxRec> ctxs_snapshot() const;
  void track_module(CUmodule m, CUcontext ctx, const char* path, const void* image, size_t image_len);
  void untrack_module(CUmodule m);
  std::vector<ModuleRec> modules_snapshot() const;
  size_t module_count() const;
  void track_symbol(CUfunction fn, CUmodule mod, const char* name);
  std::vector<SymbolRec> symbols_for_module(CUmodule mod) const;
  size_t symbol_count() const;
  void track_event(CUevent e, CUcontext ctx, unsigned fl);
  void untrack_event(CUevent e);
  std::vector<EventRec> events_snapshot() const;
  size_t event_count() const;
  void host_callback_enter();
  void host_callback_leave();
  size_t pending_host_callbacks() const;
  void set_primary(CUcontext c);
  CUcontext primary() const;
  unsigned primary_flags() const;
  void set_device(CUdevice d);
  CUdevice device() const;
  void set_device_flags(unsigned fl);
  unsigned device_flags() const;
  bool unsupported_detected() const;
  uint32_t unsupported_code() const;
  void mark_unsupported(uint32_t code, const char* reason);
  std::string unsupported_reason() const;
  TrackerStats stats_snapshot() const;
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
  mutable std::mutex module_mu_;
  std::unordered_map<CUmodule, ModuleRec> modules_;
  mutable std::mutex symbol_mu_;
  std::vector<SymbolRec> symbols_;
  mutable std::mutex event_mu_;
  std::unordered_map<CUevent, EventRec> events_;
  CUcontext primary_{nullptr};
  unsigned primary_flags_{0};
  CUdevice device_{0};
  unsigned device_flags_{0};
  mutable std::mutex meta_mu_;
  bool bad_{false};
  uint32_t bad_code_{0};
  std::string bad_reason_;
  std::string image_dir_;
  int state_{0};
  uint64_t next_seq_{1};
  size_t host_callbacks_{0};
  Shard& shard(CUdeviceptr p);
  const Shard& shard(CUdeviceptr p) const;
};
