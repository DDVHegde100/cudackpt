#include "errcode.h"
#include "tracker.hpp"
#include <cuda.h>
#undef cuMemAlloc
#undef cuMemFree
#undef cuMemcpyDtoH
#undef cuMemcpyHtoD
#undef cuStreamCreate
#include <dlfcn.h>
#include <pthread.h>
#include <stdio.h>
#include <string.h>

typedef CUresult (*cuInit_t)(unsigned int);
typedef CUresult (*cuMemAlloc_t)(CUdeviceptr*, size_t);
typedef CUresult (*cuMemFree_t)(CUdeviceptr);
typedef CUresult (*cuMemAlloc_v2_t)(CUdeviceptr*, size_t);
typedef CUresult (*cuMemFree_v2_t)(CUdeviceptr);
typedef CUresult (*cuCtxCreate_t)(CUcontext*, unsigned, CUdevice);
typedef CUresult (*cuCtxDestroy_t)(CUcontext);
typedef CUresult (*cuCtxSetCurrent_t)(CUcontext);
typedef CUresult (*cuCtxGetCurrent_t)(CUcontext*);
typedef CUresult (*cuStreamCreate_t)(CUstream*, unsigned);
typedef CUresult (*cuStreamCreate_v2_t)(CUstream*, unsigned);
typedef CUresult (*cuStreamDestroy_t)(CUstream);
typedef CUresult (*cuStreamSynchronize_t)(CUstream);
typedef CUresult (*cuCtxSynchronize_t)();
typedef CUresult (*cuLaunchKernel_t)(CUfunction, unsigned, unsigned, unsigned,
                                       unsigned, unsigned, unsigned, unsigned,
                                       CUstream, void**, void**);
typedef CUresult (*cuMemcpyDtoH_t)(void*, CUdeviceptr, size_t);
typedef CUresult (*cuMemcpyHtoD_t)(CUdeviceptr, const void*, size_t);
typedef CUresult (*cuMemcpyDtoH_v2_t)(void*, CUdeviceptr, size_t);
typedef CUresult (*cuMemcpyHtoD_v2_t)(CUdeviceptr, const void*, size_t);
typedef CUresult (*cuDeviceGet_t)(CUdevice*, int);
typedef CUresult (*cuDeviceGetCount_t)(int*);
typedef CUresult (*cuDevicePrimaryCtxRetain_t)(CUcontext*, CUdevice);
typedef CUresult (*cuDeviceSetFlags_t)(unsigned);
typedef CUresult (*cuGetProcAddress_t)(const char*, void**, int, cuuint64_t,
                                       CUdriverProcAddressQueryResult*);
typedef CUresult (*cuModuleLoad_t)(CUmodule*, const char*);
typedef CUresult (*cuModuleLoadData_t)(CUmodule*, const void*);
typedef CUresult (*cuModuleUnload_t)(CUmodule);
typedef CUresult (*cuModuleGetFunction_t)(CUfunction*, CUmodule, const char*);
typedef CUresult (*cuEventCreate_t)(CUevent*, unsigned);
typedef CUresult (*cuEventDestroy_t)(CUevent);
typedef CUresult (*cuEventRecord_t)(CUevent, CUstream);
typedef CUresult (*cuEventSynchronize_t)(CUevent);
typedef CUresult (*cuLaunchHostFunc_t)(CUstream, CUhostFn, void*);
typedef CUresult (*cuGraphLaunch_t)(CUgraphExec, CUstream);
typedef CUresult (*cuStreamBeginCapture_t)(CUstream, CUstreamCaptureMode);
typedef CUresult (*cuStreamEndCapture_t)(CUstream, CUgraph*);
typedef CUresult (*cuDeviceCanAccessPeer_t)(int*, CUdevice, CUdevice);

static pthread_once_t once = PTHREAD_ONCE_INIT;
static cuInit_t real_cuInit;
static cuMemAlloc_t real_cuMemAlloc;
static cuMemFree_t real_cuMemFree;
static cuMemAlloc_v2_t real_cuMemAlloc_v2;
static cuMemFree_v2_t real_cuMemFree_v2;
static cuCtxCreate_t real_cuCtxCreate;
static cuCtxDestroy_t real_cuCtxDestroy;
static cuCtxSetCurrent_t real_cuCtxSetCurrent;
static cuCtxGetCurrent_t real_cuCtxGetCurrent;
static cuStreamCreate_t real_cuStreamCreate;
static cuStreamCreate_v2_t real_cuStreamCreate_v2;
static cuStreamDestroy_t real_cuStreamDestroy;
static cuStreamSynchronize_t real_cuStreamSynchronize;
static cuCtxSynchronize_t real_cuCtxSynchronize;
static cuLaunchKernel_t real_cuLaunchKernel;
static cuMemcpyDtoH_t real_cuMemcpyDtoH;
static cuMemcpyHtoD_t real_cuMemcpyHtoD;
static cuMemcpyDtoH_v2_t real_cuMemcpyDtoH_v2;
static cuMemcpyHtoD_v2_t real_cuMemcpyHtoD_v2;
static cuDeviceGet_t real_cuDeviceGet;
static cuDeviceGetCount_t real_cuDeviceGetCount;
static cuDevicePrimaryCtxRetain_t real_cuDevicePrimaryCtxRetain;
static cuDeviceSetFlags_t real_cuDeviceSetFlags;
static cuGetProcAddress_t real_cuGetProcAddress;
static cuModuleLoad_t real_cuModuleLoad;
static cuModuleLoadData_t real_cuModuleLoadData;
static cuModuleUnload_t real_cuModuleUnload;
static cuModuleGetFunction_t real_cuModuleGetFunction;
static cuEventCreate_t real_cuEventCreate;
static cuEventDestroy_t real_cuEventDestroy;
static cuEventRecord_t real_cuEventRecord;
static cuEventSynchronize_t real_cuEventSynchronize;
static cuLaunchHostFunc_t real_cuLaunchHostFunc;
static cuGraphLaunch_t real_cuGraphLaunch;
static cuStreamBeginCapture_t real_cuStreamBeginCapture;
static cuStreamEndCapture_t real_cuStreamEndCapture;
static cuDeviceCanAccessPeer_t real_cuDeviceCanAccessPeer;

static void load_syms() {
  real_cuInit = (cuInit_t)dlsym(RTLD_NEXT, "cuInit");
  real_cuMemAlloc = (cuMemAlloc_t)dlsym(RTLD_NEXT, "cuMemAlloc");
  real_cuMemFree = (cuMemFree_t)dlsym(RTLD_NEXT, "cuMemFree");
  real_cuMemAlloc_v2 = (cuMemAlloc_v2_t)dlsym(RTLD_NEXT, "cuMemAlloc_v2");
  real_cuMemFree_v2 = (cuMemFree_v2_t)dlsym(RTLD_NEXT, "cuMemFree_v2");
  real_cuCtxCreate = (cuCtxCreate_t)dlsym(RTLD_NEXT, "cuCtxCreate");
  real_cuCtxDestroy = (cuCtxDestroy_t)dlsym(RTLD_NEXT, "cuCtxDestroy");
  real_cuCtxSetCurrent = (cuCtxSetCurrent_t)dlsym(RTLD_NEXT, "cuCtxSetCurrent");
  real_cuCtxGetCurrent = (cuCtxGetCurrent_t)dlsym(RTLD_NEXT, "cuCtxGetCurrent");
  real_cuStreamCreate = (cuStreamCreate_t)dlsym(RTLD_NEXT, "cuStreamCreate");
  real_cuStreamCreate_v2 = (cuStreamCreate_v2_t)dlsym(RTLD_NEXT, "cuStreamCreate_v2");
  real_cuStreamDestroy = (cuStreamDestroy_t)dlsym(RTLD_NEXT, "cuStreamDestroy");
  real_cuStreamSynchronize =
      (cuStreamSynchronize_t)dlsym(RTLD_NEXT, "cuStreamSynchronize");
  real_cuCtxSynchronize = (cuCtxSynchronize_t)dlsym(RTLD_NEXT, "cuCtxSynchronize");
  real_cuLaunchKernel = (cuLaunchKernel_t)dlsym(RTLD_NEXT, "cuLaunchKernel");
  real_cuMemcpyDtoH = (cuMemcpyDtoH_t)dlsym(RTLD_NEXT, "cuMemcpyDtoH");
  real_cuMemcpyHtoD = (cuMemcpyHtoD_t)dlsym(RTLD_NEXT, "cuMemcpyHtoD");
  real_cuMemcpyDtoH_v2 = (cuMemcpyDtoH_v2_t)dlsym(RTLD_NEXT, "cuMemcpyDtoH_v2");
  real_cuMemcpyHtoD_v2 = (cuMemcpyHtoD_v2_t)dlsym(RTLD_NEXT, "cuMemcpyHtoD_v2");
  real_cuDeviceGet = (cuDeviceGet_t)dlsym(RTLD_NEXT, "cuDeviceGet");
  real_cuDeviceGetCount = (cuDeviceGetCount_t)dlsym(RTLD_NEXT, "cuDeviceGetCount");
  real_cuDevicePrimaryCtxRetain =
      (cuDevicePrimaryCtxRetain_t)dlsym(RTLD_NEXT, "cuDevicePrimaryCtxRetain");
  real_cuDeviceSetFlags = (cuDeviceSetFlags_t)dlsym(RTLD_NEXT, "cuDeviceSetFlags");
  real_cuGetProcAddress =
      (cuGetProcAddress_t)dlsym(RTLD_NEXT, "cuGetProcAddress");
  real_cuModuleLoad = (cuModuleLoad_t)dlsym(RTLD_NEXT, "cuModuleLoad");
  real_cuModuleLoadData = (cuModuleLoadData_t)dlsym(RTLD_NEXT, "cuModuleLoadData");
  real_cuModuleUnload = (cuModuleUnload_t)dlsym(RTLD_NEXT, "cuModuleUnload");
  real_cuModuleGetFunction =
      (cuModuleGetFunction_t)dlsym(RTLD_NEXT, "cuModuleGetFunction");
  real_cuEventCreate = (cuEventCreate_t)dlsym(RTLD_NEXT, "cuEventCreate");
  real_cuEventDestroy = (cuEventDestroy_t)dlsym(RTLD_NEXT, "cuEventDestroy");
  real_cuEventRecord = (cuEventRecord_t)dlsym(RTLD_NEXT, "cuEventRecord");
  real_cuEventSynchronize = (cuEventSynchronize_t)dlsym(RTLD_NEXT, "cuEventSynchronize");
  real_cuLaunchHostFunc = (cuLaunchHostFunc_t)dlsym(RTLD_NEXT, "cuLaunchHostFunc");
  real_cuGraphLaunch = (cuGraphLaunch_t)dlsym(RTLD_NEXT, "cuGraphLaunch");
  real_cuStreamBeginCapture =
      (cuStreamBeginCapture_t)dlsym(RTLD_NEXT, "cuStreamBeginCapture");
  real_cuStreamEndCapture = (cuStreamEndCapture_t)dlsym(RTLD_NEXT, "cuStreamEndCapture");
  real_cuDeviceCanAccessPeer =
      (cuDeviceCanAccessPeer_t)dlsym(RTLD_NEXT, "cuDeviceCanAccessPeer");
}

static void guard_symbol(const char* api) {
  if (!api) return;
  if (strstr(api, "nccl") || strstr(api, "NCCL"))
    Tracker::instance().mark_unsupported(CKPT_ERR_NCCL, "nccl");
  if (strstr(api, "Graph") || strstr(api, "graph"))
    Tracker::instance().mark_unsupported(CKPT_ERR_CUDA_GRAPH, "cuda graph");
  if (strstr(api, "MIG") || strstr(api, "mig"))
    Tracker::instance().mark_unsupported(CKPT_ERR_MIG, "mig");
}

struct HostWrap {
  CUhostFn fn;
  void* data;
};

static void CKAPI host_wrap(void* arg) {
  auto* w = static_cast<HostWrap*>(arg);
  w->fn(w->data);
  Tracker::instance().host_callback_leave();
  delete w;
}

extern "C" {

CUresult cuInit(unsigned int flags) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuInit(flags);
  if (r == CUDA_SUCCESS && real_cuDeviceGetCount) {
    int n = 0;
    if (real_cuDeviceGetCount(&n) == CUDA_SUCCESS && n > 1)
      Tracker::instance().mark_unsupported(CKPT_ERR_MULTI_GPU, "multi gpu");
  }
  return r;
}

CUresult cuDeviceGet(CUdevice* dev, int ord) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuDeviceGet(dev, ord);
  if (r == CUDA_SUCCESS && dev) Tracker::instance().set_device(*dev);
  return r;
}

CUresult cuDeviceSetFlags(unsigned flags) {
  pthread_once(&once, load_syms);
  Tracker::instance().set_device_flags(flags);
  if (real_cuDeviceSetFlags) return real_cuDeviceSetFlags(flags);
  return CUDA_SUCCESS;
}

CUresult cuDevicePrimaryCtxRetain(CUcontext* ctx, CUdevice dev) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuDevicePrimaryCtxRetain(ctx, dev);
  if (r == CUDA_SUCCESS && ctx && *ctx) {
    Tracker::instance().track_ctx(*ctx, dev, 0);
    Tracker::instance().set_primary(*ctx);
  }
  return r;
}

CUresult cuCtxCreate(CUcontext* ctx, unsigned int flags, CUdevice dev) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuCtxCreate(ctx, flags, dev);
  if (r == CUDA_SUCCESS && ctx && *ctx) {
    Tracker::instance().track_ctx(*ctx, dev, flags);
    Tracker::instance().set_primary(*ctx);
  }
  return r;
}

CUresult cuCtxDestroy(CUcontext ctx) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_ctx(ctx);
  return real_cuCtxDestroy(ctx);
}

CUresult cuCtxSetCurrent(CUcontext ctx) {
  pthread_once(&once, load_syms);
  if (ctx) Tracker::instance().set_primary(ctx);
  return real_cuCtxSetCurrent(ctx);
}

CUresult cuCtxGetCurrent(CUcontext* ctx) {
  pthread_once(&once, load_syms);
  return real_cuCtxGetCurrent(ctx);
}

CUresult cuModuleLoad(CUmodule* mod, const char* path) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuModuleLoad(mod, path);
  if (r == CUDA_SUCCESS && mod && *mod)
    Tracker::instance().track_module(*mod, ctx, path, nullptr, 0);
  return r;
}

CUresult cuModuleLoadData(CUmodule* mod, const void* image) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuModuleLoadData(mod, image);
  if (r == CUDA_SUCCESS && mod && *mod)
    Tracker::instance().track_module(*mod, ctx, "", image, 0);
  return r;
}

CUresult cuModuleUnload(CUmodule mod) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_module(mod);
  return real_cuModuleUnload(mod);
}

CUresult cuModuleGetFunction(CUfunction* fn, CUmodule mod, const char* name) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuModuleGetFunction(fn, mod, name);
  if (r == CUDA_SUCCESS && fn && *fn) Tracker::instance().track_symbol(*fn, mod, name);
  return r;
}

CUresult cuMemAlloc(CUdeviceptr* dptr, size_t bytesize) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuMemAlloc(dptr, bytesize);
  if (r == CUDA_SUCCESS && dptr)
    Tracker::instance().track_alloc(*dptr, bytesize, ctx, 0);
  return r;
}

CUresult cuMemAlloc_v2(CUdeviceptr* dptr, size_t bytesize) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuMemAlloc_v2 ? real_cuMemAlloc_v2(dptr, bytesize)
                                  : real_cuMemAlloc(dptr, bytesize);
  if (r == CUDA_SUCCESS && dptr)
    Tracker::instance().track_alloc(*dptr, bytesize, ctx, 0);
  return r;
}

CUresult cuMemFree(CUdeviceptr dptr) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_alloc(dptr);
  return real_cuMemFree(dptr);
}

CUresult cuMemFree_v2(CUdeviceptr dptr) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_alloc(dptr);
  if (real_cuMemFree_v2) return real_cuMemFree_v2(dptr);
  return real_cuMemFree(dptr);
}

CUresult cuStreamCreate(CUstream* phStream, unsigned int flags) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuStreamCreate(phStream, flags);
  if (r == CUDA_SUCCESS && phStream)
    Tracker::instance().track_stream(*phStream, ctx, flags);
  return r;
}

CUresult cuStreamCreate_v2(CUstream* phStream, unsigned int flags) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuStreamCreate_v2 ? real_cuStreamCreate_v2(phStream, flags)
                                      : real_cuStreamCreate(phStream, flags);
  if (r == CUDA_SUCCESS && phStream)
    Tracker::instance().track_stream(*phStream, ctx, flags);
  return r;
}

CUresult cuStreamDestroy(CUstream hStream) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_stream(hStream);
  return real_cuStreamDestroy(hStream);
}

CUresult cuStreamSynchronize(CUstream hStream) {
  pthread_once(&once, load_syms);
  return real_cuStreamSynchronize(hStream);
}

CUresult cuCtxSynchronize() {
  pthread_once(&once, load_syms);
  if (real_cuCtxSynchronize) return real_cuCtxSynchronize();
  return CUDA_SUCCESS;
}

CUresult cuEventCreate(CUevent* ev, unsigned flags) {
  pthread_once(&once, load_syms);
  CUcontext ctx = nullptr;
  real_cuCtxGetCurrent(&ctx);
  CUresult r = real_cuEventCreate(ev, flags);
  if (r == CUDA_SUCCESS && ev && *ev) Tracker::instance().track_event(*ev, ctx, flags);
  return r;
}

CUresult cuEventDestroy(CUevent ev) {
  pthread_once(&once, load_syms);
  Tracker::instance().untrack_event(ev);
  return real_cuEventDestroy(ev);
}

CUresult cuEventRecord(CUevent ev, CUstream stream) {
  pthread_once(&once, load_syms);
  return real_cuEventRecord(ev, stream);
}

CUresult cuEventSynchronize(CUevent ev) {
  pthread_once(&once, load_syms);
  return real_cuEventSynchronize(ev);
}

CUresult cuLaunchHostFunc(CUstream stream, CUhostFn fn, void* data) {
  pthread_once(&once, load_syms);
  if (!real_cuLaunchHostFunc) return CUDA_ERROR_NOT_SUPPORTED;
  auto* wrap = new HostWrap{fn, data};
  Tracker::instance().host_callback_enter();
  return real_cuLaunchHostFunc(stream, host_wrap, wrap);
}

CUresult cuLaunchKernel(CUfunction f, unsigned int gx, unsigned int gy,
                        unsigned int gz, unsigned int bx, unsigned int by,
                        unsigned int bz, unsigned int shmem, CUstream st,
                        void** args, void** extra) {
  pthread_once(&once, load_syms);
  return real_cuLaunchKernel(f, gx, gy, gz, bx, by, bz, shmem, st, args, extra);
}

CUresult cuGraphLaunch(CUgraphExec exec, CUstream stream) {
  pthread_once(&once, load_syms);
  Tracker::instance().mark_unsupported(CKPT_ERR_CUDA_GRAPH, "cuda graph");
  (void)exec;
  (void)stream;
  return CUDA_ERROR_NOT_SUPPORTED;
}

CUresult cuStreamBeginCapture(CUstream stream, CUstreamCaptureMode mode) {
  pthread_once(&once, load_syms);
  Tracker::instance().mark_unsupported(CKPT_ERR_CUDA_GRAPH, "cuda graph");
  (void)stream;
  (void)mode;
  return CUDA_ERROR_NOT_SUPPORTED;
}

CUresult cuStreamEndCapture(CUstream stream, CUgraph* graph) {
  pthread_once(&once, load_syms);
  Tracker::instance().mark_unsupported(CKPT_ERR_CUDA_GRAPH, "cuda graph");
  (void)stream;
  (void)graph;
  return CUDA_ERROR_NOT_SUPPORTED;
}

CUresult cuDeviceCanAccessPeer(int* can, CUdevice dev, CUdevice peer) {
  pthread_once(&once, load_syms);
  if (dev != peer) Tracker::instance().mark_unsupported(CKPT_ERR_PEER_ACCESS, "peer access");
  if (real_cuDeviceCanAccessPeer) return real_cuDeviceCanAccessPeer(can, dev, peer);
  return CUDA_ERROR_NOT_SUPPORTED;
}

CUresult cuMemcpyDtoH(void* dst, CUdeviceptr src, size_t n) {
  pthread_once(&once, load_syms);
  return real_cuMemcpyDtoH(dst, src, n);
}

CUresult cuMemcpyHtoD(CUdeviceptr dst, const void* src, size_t n) {
  pthread_once(&once, load_syms);
  return real_cuMemcpyHtoD(dst, src, n);
}

CUresult cuMemcpyDtoH_v2(void* dst, CUdeviceptr src, size_t n) {
  pthread_once(&once, load_syms);
  if (real_cuMemcpyDtoH_v2) return real_cuMemcpyDtoH_v2(dst, src, n);
  return real_cuMemcpyDtoH(dst, src, n);
}

CUresult cuMemcpyHtoD_v2(CUdeviceptr dst, const void* src, size_t n) {
  pthread_once(&once, load_syms);
  if (real_cuMemcpyHtoD_v2) return real_cuMemcpyHtoD_v2(dst, src, n);
  return real_cuMemcpyHtoD(dst, src, n);
}

CUresult cuGetProcAddress_v2(const char* symbol, void** pfn, int cudaVersion,
                             cuuint64_t flags,
                             CUdriverProcAddressQueryResult* symbolStatus) {
  pthread_once(&once, load_syms);
  guard_symbol(symbol ? symbol : "");
  if (real_cuGetProcAddress)
    return real_cuGetProcAddress(symbol, pfn, cudaVersion, flags, symbolStatus);
  return CUDA_ERROR_NOT_FOUND;
}

}

extern "C" int ckpt_wait_frozen();
extern "C" int ckpt_app_resumed();
