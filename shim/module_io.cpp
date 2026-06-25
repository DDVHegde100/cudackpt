#include "module_io.hpp"
#include "log.h"
#include "tracker.hpp"
#include <cuda.h>
#include <fstream>
#include <string>
#include <vector>

static const uint32_t kModMagic = 0x4D4F4455u;
static const uint32_t kCtxMagic = 0x43585458u;
static const uint32_t kStrMagic = 0x5354524Du;

static std::string path_join(const char* dir, const char* name) {
  std::string s(dir);
  if (!s.empty() && s.back() != '/') s += '/';
  s += name;
  return s;
}

static void write_u32(std::ofstream& f, uint32_t v) { f.write(reinterpret_cast<const char*>(&v), 4); }

int ckpt_modules_write(const char* dir) {
  auto mods = Tracker::instance().modules_snapshot();
  std::ofstream f(path_join(dir, "modules.bin"), std::ios::binary | std::ios::trunc);
  if (!f) return -30;
  write_u32(f, kModMagic);
  uint16_t ver = 1;
  f.write(reinterpret_cast<const char*>(&ver), 2);
  uint16_t pad = 0;
  f.write(reinterpret_cast<const char*>(&pad), 2);
  write_u32(f, static_cast<uint32_t>(mods.size()));
  for (const auto& m : mods) {
    uint64_t ctx = reinterpret_cast<uint64_t>(m.ctx);
    f.write(reinterpret_cast<const char*>(&ctx), 8);
    uint32_t kind = m.image.empty() ? 0u : 1u;
    write_u32(f, kind);
    if (kind == 0) {
      write_u32(f, static_cast<uint32_t>(m.path.size()));
      if (!m.path.empty()) f.write(m.path.data(), static_cast<std::streamsize>(m.path.size()));
    } else {
      write_u32(f, static_cast<uint32_t>(m.image.size()));
      if (!m.image.empty())
        f.write(reinterpret_cast<const char*>(m.image.data()),
                static_cast<std::streamsize>(m.image.size()));
    }
    auto syms = Tracker::instance().symbols_for_module(m.mod);
    write_u32(f, static_cast<uint32_t>(syms.size()));
    for (const auto& s : syms) {
      write_u32(f, static_cast<uint32_t>(s.name.size()));
      if (!s.name.empty()) f.write(s.name.data(), static_cast<std::streamsize>(s.name.size()));
    }
  }
  return 0;
}

int ckpt_modules_restore(const char* dir) {
  std::ifstream f(path_join(dir, "modules.bin"), std::ios::binary);
  if (!f) return 0;
  uint32_t magic = 0;
  f.read(reinterpret_cast<char*>(&magic), 4);
  if (magic != kModMagic) return -31;
  uint16_t ver = 0;
  f.read(reinterpret_cast<char*>(&ver), 2);
  f.seekg(2, std::ios::cur);
  uint32_t count = 0;
  f.read(reinterpret_cast<char*>(&count), 4);
  auto& tr = Tracker::instance();
  CUcontext cur = tr.primary();
  if (cur) cuCtxSetCurrent(cur);
  for (uint32_t i = 0; i < count; ++i) {
    uint64_t ctx_u = 0;
    f.read(reinterpret_cast<char*>(&ctx_u), 8);
    CUcontext ctx = reinterpret_cast<CUcontext>(ctx_u);
    if (ctx) cuCtxSetCurrent(ctx);
    uint32_t kind = 0;
    f.read(reinterpret_cast<char*>(&kind), 4);
    CUmodule mod = nullptr;
    if (kind == 0) {
      uint32_t plen = 0;
      f.read(reinterpret_cast<char*>(&plen), 4);
      std::string path(plen, '\0');
      if (plen) f.read(path.data(), plen);
      CUresult r = cuModuleLoad(&mod, path.c_str());
      if (r != CUDA_SUCCESS) return -32;
      tr.track_module(mod, ctx, path.c_str(), nullptr, 0);
    } else {
      uint32_t ilen = 0;
      f.read(reinterpret_cast<char*>(&ilen), 4);
      std::vector<uint8_t> image(ilen);
      if (ilen) f.read(reinterpret_cast<char*>(image.data()), ilen);
      CUresult r = cuModuleLoadData(&mod, image.data());
      if (r != CUDA_SUCCESS) return -33;
      tr.track_module(mod, ctx, "", image.data(), image.size());
    }
    uint32_t sc = 0;
    f.read(reinterpret_cast<char*>(&sc), 4);
    for (uint32_t s = 0; s < sc; ++s) {
      uint32_t nlen = 0;
      f.read(reinterpret_cast<char*>(&nlen), 4);
      std::string name(nlen, '\0');
      if (nlen) f.read(name.data(), nlen);
      CUfunction fn = nullptr;
      if (cuModuleGetFunction(&fn, mod, name.c_str()) == CUDA_SUCCESS)
        tr.track_symbol(fn, mod, name.c_str());
    }
  }
  return 0;
}

int ckpt_ctx_write(const char* dir) {
  auto& tr = Tracker::instance();
  std::ofstream f(path_join(dir, "ctx.bin"), std::ios::binary | std::ios::trunc);
  if (!f) return -40;
  write_u32(f, kCtxMagic);
  uint16_t ver = 1;
  f.write(reinterpret_cast<const char*>(&ver), 2);
  uint16_t pad = 0;
  f.write(reinterpret_cast<const char*>(&pad), 2);
  uint64_t primary = reinterpret_cast<uint64_t>(tr.primary());
  f.write(reinterpret_cast<const char*>(&primary), 8);
  uint32_t dev = static_cast<uint32_t>(tr.device());
  write_u32(f, dev);
  uint32_t ctx_flags = tr.primary_flags();
  write_u32(f, ctx_flags);
  uint32_t dev_flags = tr.device_flags();
  write_u32(f, dev_flags);
  return 0;
}

int ckpt_ctx_restore(const char* dir) {
  std::ifstream f(path_join(dir, "ctx.bin"), std::ios::binary);
  if (!f) return 0;
  uint32_t magic = 0;
  f.read(reinterpret_cast<char*>(&magic), 4);
  if (magic != kCtxMagic) return 0;
  f.seekg(4, std::ios::cur);
  uint64_t primary = 0;
  f.read(reinterpret_cast<char*>(&primary), 8);
  uint32_t dev = 0, ctx_flags = 0, dev_flags = 0;
  f.read(reinterpret_cast<char*>(&dev), 4);
  f.read(reinterpret_cast<char*>(&ctx_flags), 4);
  f.read(reinterpret_cast<char*>(&dev_flags), 4);
  auto& tr = Tracker::instance();
  tr.set_device(static_cast<CUdevice>(dev));
  tr.set_device_flags(dev_flags);
  CUcontext ctx = reinterpret_cast<CUcontext>(primary);
  if (ctx) {
    tr.set_primary(ctx);
    tr.track_ctx(ctx, static_cast<CUdevice>(dev), ctx_flags);
    cuCtxSetCurrent(ctx);
  }
  return 0;
}

int ckpt_streams_write(const char* dir) {
  auto streams = Tracker::instance().streams_snapshot();
  std::ofstream f(path_join(dir, "streams.bin"), std::ios::binary | std::ios::trunc);
  if (!f) return -50;
  write_u32(f, kStrMagic);
  uint16_t ver = 1;
  f.write(reinterpret_cast<const char*>(&ver), 2);
  uint16_t pad = 0;
  f.write(reinterpret_cast<const char*>(&pad), 2);
  write_u32(f, static_cast<uint32_t>(streams.size()));
  for (const auto& sr : streams) {
    int pri = sr.priority;
    cuStreamGetPriority(sr.stream, &pri);
    unsigned cap = sr.capture_status;
    CUstreamCaptureStatus st = CU_STREAM_CAPTURE_STATUS_NONE;
    if (cuStreamIsCapturing(sr.stream, &st) == CUDA_SUCCESS)
      cap = static_cast<unsigned>(st);
    uint64_t stv = reinterpret_cast<uint64_t>(sr.stream);
    uint64_t ctx = reinterpret_cast<uint64_t>(sr.ctx);
    f.write(reinterpret_cast<const char*>(&stv), 8);
    f.write(reinterpret_cast<const char*>(&ctx), 8);
    write_u32(f, sr.flags);
    int32_t ipri = pri;
    f.write(reinterpret_cast<const char*>(&ipri), 4);
    write_u32(f, cap);
  }
  return 0;
}

int ckpt_streams_restore(const char* dir) {
  std::ifstream f(path_join(dir, "streams.bin"), std::ios::binary);
  if (!f) return 0;
  uint32_t magic = 0;
  f.read(reinterpret_cast<char*>(&magic), 4);
  if (magic != kStrMagic) return 0;
  f.seekg(4, std::ios::cur);
  uint32_t count = 0;
  f.read(reinterpret_cast<char*>(&count), 4);
  for (uint32_t i = 0; i < count; ++i) {
    uint64_t st = 0, ctx = 0;
    f.read(reinterpret_cast<char*>(&st), 8);
    f.read(reinterpret_cast<char*>(&ctx), 8);
    uint32_t flags = 0, capture = 0;
    int32_t pri = 0;
    f.read(reinterpret_cast<char*>(&flags), 4);
    f.read(reinterpret_cast<char*>(&pri), 4);
    f.read(reinterpret_cast<char*>(&capture), 4);
    CUstream stream = reinterpret_cast<CUstream>(st);
    Tracker::instance().update_stream_state(stream, pri, capture);
  }
  return 0;
}
