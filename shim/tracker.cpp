#include "tracker.hpp"
#include <algorithm>

Tracker& Tracker::instance() {
  static Tracker t;
  return t;
}

Tracker::Shard& Tracker::shard(CUdeviceptr p) {
  return shards_[static_cast<size_t>(p >> 6) % kShards];
}

const Tracker::Shard& Tracker::shard(CUdeviceptr p) const {
  return shards_[static_cast<size_t>(p >> 6) % kShards];
}

void Tracker::track_alloc(CUdeviceptr p, size_t sz, CUcontext ctx, unsigned fl) {
  if (!p) return;
  uint64_t seq = 0;
  {
    std::lock_guard<std::mutex> lk(meta_mu_);
    seq = next_seq_++;
  }
  auto& s = shard(p);
  std::lock_guard<std::mutex> lk(s.mu);
  s.map[p] = AllocRec{p, sz, ctx, fl, seq};
}

void Tracker::untrack_alloc(CUdeviceptr p) {
  if (!p) return;
  auto& s = shard(p);
  std::lock_guard<std::mutex> lk(s.mu);
  s.map.erase(p);
}

bool Tracker::lookup_alloc(CUdeviceptr p, AllocRec* out) const {
  auto& s = shard(p);
  std::lock_guard<std::mutex> lk(s.mu);
  auto it = s.map.find(p);
  if (it == s.map.end()) return false;
  if (out) *out = it->second;
  return true;
}

std::vector<AllocRec> Tracker::allocs_snapshot() const {
  std::vector<AllocRec> v;
  v.reserve(256);
  for (int i = 0; i < kShards; ++i) {
    std::lock_guard<std::mutex> lk(shards_[i].mu);
    for (const auto& kv : shards_[i].map) v.push_back(kv.second);
  }
  std::sort(v.begin(), v.end(),
            [](const AllocRec& a, const AllocRec& b) { return a.seq < b.seq; });
  return v;
}

size_t Tracker::alloc_count() const {
  size_t n = 0;
  for (int i = 0; i < kShards; ++i) {
    std::lock_guard<std::mutex> lk(shards_[i].mu);
    n += shards_[i].map.size();
  }
  return n;
}

uint64_t Tracker::total_bytes() const {
  uint64_t t = 0;
  for (int i = 0; i < kShards; ++i) {
    std::lock_guard<std::mutex> lk(shards_[i].mu);
    for (const auto& kv : shards_[i].map) t += static_cast<uint64_t>(kv.second.size);
  }
  return t;
}

void Tracker::track_stream(CUstream s, CUcontext ctx, unsigned fl) {
  if (!s) return;
  std::lock_guard<std::mutex> lk(stream_mu_);
  streams_[s] = StreamRec{s, ctx, fl, 0, 0};
}

void Tracker::untrack_stream(CUstream s) {
  if (!s) return;
  std::lock_guard<std::mutex> lk(stream_mu_);
  streams_.erase(s);
}

void Tracker::update_stream_state(CUstream s, int priority, unsigned capture_status) {
  if (!s) return;
  std::lock_guard<std::mutex> lk(stream_mu_);
  auto it = streams_.find(s);
  if (it != streams_.end()) {
    it->second.priority = priority;
    it->second.capture_status = capture_status;
  }
}

std::vector<StreamRec> Tracker::streams_snapshot() const {
  std::lock_guard<std::mutex> lk(stream_mu_);
  std::vector<StreamRec> v;
  v.reserve(streams_.size());
  for (const auto& kv : streams_) v.push_back(kv.second);
  return v;
}

void Tracker::track_ctx(CUcontext c, CUdevice d, unsigned fl) {
  if (!c) return;
  std::lock_guard<std::mutex> lk(ctx_mu_);
  ctxs_[c] = CtxRec{c, d, fl};
}

void Tracker::untrack_ctx(CUcontext c) {
  if (!c) return;
  std::lock_guard<std::mutex> lk(ctx_mu_);
  ctxs_.erase(c);
}

std::vector<CtxRec> Tracker::ctxs_snapshot() const {
  std::lock_guard<std::mutex> lk(ctx_mu_);
  std::vector<CtxRec> v;
  v.reserve(ctxs_.size());
  for (const auto& kv : ctxs_) v.push_back(kv.second);
  return v;
}

void Tracker::track_module(CUmodule m, CUcontext ctx, const char* path, const void* image,
                           size_t image_len) {
  if (!m) return;
  uint64_t seq = 0;
  {
    std::lock_guard<std::mutex> lk(meta_mu_);
    seq = next_seq_++;
  }
  ModuleRec rec{m, ctx, path ? path : "", {}, seq};
  if (image && image_len > 0) {
    rec.image.assign(static_cast<const uint8_t*>(image),
                     static_cast<const uint8_t*>(image) + image_len);
  }
  std::lock_guard<std::mutex> lk(module_mu_);
  modules_[m] = std::move(rec);
}

void Tracker::untrack_module(CUmodule m) {
  if (!m) return;
  std::lock_guard<std::mutex> lk(module_mu_);
  modules_.erase(m);
}

std::vector<ModuleRec> Tracker::modules_snapshot() const {
  std::lock_guard<std::mutex> lk(module_mu_);
  std::vector<ModuleRec> v;
  v.reserve(modules_.size());
  for (const auto& kv : modules_) v.push_back(kv.second);
  std::sort(v.begin(), v.end(),
            [](const ModuleRec& a, const ModuleRec& b) { return a.seq < b.seq; });
  return v;
}

size_t Tracker::module_count() const {
  std::lock_guard<std::mutex> lk(module_mu_);
  return modules_.size();
}

void Tracker::track_symbol(CUfunction fn, CUmodule mod, const char* name) {
  if (!fn || !name) return;
  std::lock_guard<std::mutex> lk(symbol_mu_);
  symbols_.push_back(SymbolRec{fn, mod, name});
}

std::vector<SymbolRec> Tracker::symbols_for_module(CUmodule mod) const {
  std::lock_guard<std::mutex> lk(symbol_mu_);
  std::vector<SymbolRec> v;
  for (const auto& s : symbols_) {
    if (s.mod == mod) v.push_back(s);
  }
  return v;
}

size_t Tracker::symbol_count() const {
  std::lock_guard<std::mutex> lk(symbol_mu_);
  return symbols_.size();
}

void Tracker::track_event(CUevent e, CUcontext ctx, unsigned fl) {
  if (!e) return;
  std::lock_guard<std::mutex> lk(event_mu_);
  events_[e] = EventRec{e, ctx, fl};
}

void Tracker::untrack_event(CUevent e) {
  if (!e) return;
  std::lock_guard<std::mutex> lk(event_mu_);
  events_.erase(e);
}

std::vector<EventRec> Tracker::events_snapshot() const {
  std::lock_guard<std::mutex> lk(event_mu_);
  std::vector<EventRec> v;
  v.reserve(events_.size());
  for (const auto& kv : events_) v.push_back(kv.second);
  return v;
}

size_t Tracker::event_count() const {
  std::lock_guard<std::mutex> lk(event_mu_);
  return events_.size();
}

void Tracker::host_callback_enter() {
  std::lock_guard<std::mutex> lk(meta_mu_);
  ++host_callbacks_;
}

void Tracker::host_callback_leave() {
  std::lock_guard<std::mutex> lk(meta_mu_);
  if (host_callbacks_ > 0) --host_callbacks_;
}

size_t Tracker::pending_host_callbacks() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return host_callbacks_;
}

void Tracker::set_primary(CUcontext c) {
  unsigned fl = 0;
  {
    std::lock_guard<std::mutex> clk(ctx_mu_);
    auto it = ctxs_.find(c);
    if (it != ctxs_.end()) fl = it->second.flags;
  }
  std::lock_guard<std::mutex> lk(meta_mu_);
  primary_ = c;
  primary_flags_ = fl;
}

CUcontext Tracker::primary() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return primary_;
}

unsigned Tracker::primary_flags() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return primary_flags_;
}

void Tracker::set_device(CUdevice d) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  device_ = d;
}

CUdevice Tracker::device() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return device_;
}

void Tracker::set_device_flags(unsigned fl) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  device_flags_ = fl;
}

unsigned Tracker::device_flags() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return device_flags_;
}

bool Tracker::unsupported_detected() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return bad_;
}

uint32_t Tracker::unsupported_code() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return bad_code_;
}

void Tracker::mark_unsupported(uint32_t code, const char* reason) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  bad_ = true;
  bad_code_ = code;
  bad_reason_ = reason ? reason : "unsupported";
}

std::string Tracker::unsupported_reason() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return bad_reason_;
}

TrackerStats Tracker::stats_snapshot() const {
  TrackerStats s{};
  s.alloc_count = alloc_count();
  s.total_bytes = total_bytes();
  s.stream_count = streams_snapshot().size();
  s.module_count = module_count();
  s.symbol_count = symbol_count();
  s.event_count = event_count();
  s.ctx_count = ctxs_snapshot().size();
  s.unsupported_code = unsupported_code();
  s.state = state();
  return s;
}

void Tracker::clear() {
  for (int i = 0; i < kShards; ++i) {
    std::lock_guard<std::mutex> lk(shards_[i].mu);
    shards_[i].map.clear();
  }
  {
    std::lock_guard<std::mutex> lk(stream_mu_);
    streams_.clear();
  }
  {
    std::lock_guard<std::mutex> lk(ctx_mu_);
    ctxs_.clear();
  }
  {
    std::lock_guard<std::mutex> lk(module_mu_);
    modules_.clear();
  }
  {
    std::lock_guard<std::mutex> lk(symbol_mu_);
    symbols_.clear();
  }
  {
    std::lock_guard<std::mutex> lk(event_mu_);
    events_.clear();
  }
}

void Tracker::set_image_dir(const std::string& d) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  image_dir_ = d;
}

std::string Tracker::image_dir() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return image_dir_;
}

void Tracker::set_state(int s) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  state_ = s;
}

int Tracker::state() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return state_;
}
