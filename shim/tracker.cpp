#include "tracker.hpp"
#include <algorithm>
#include <cstring>

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
  streams_[s] = StreamRec{s, ctx, fl};
}

void Tracker::untrack_stream(CUstream s) {
  if (!s) return;
  std::lock_guard<std::mutex> lk(stream_mu_);
  streams_.erase(s);
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

void Tracker::set_primary(CUcontext c) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  primary_ = c;
}

CUcontext Tracker::primary() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return primary_;
}

void Tracker::set_device(CUdevice d) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  device_ = d;
}

CUdevice Tracker::device() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return device_;
}

bool Tracker::unsupported_detected() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return bad_;
}

void Tracker::mark_unsupported(const char* reason) {
  std::lock_guard<std::mutex> lk(meta_mu_);
  bad_ = true;
  bad_reason_ = reason ? reason : "unsupported";
}

std::string Tracker::unsupported_reason() const {
  std::lock_guard<std::mutex> lk(meta_mu_);
  return bad_reason_;
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
