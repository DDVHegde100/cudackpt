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
typedef CUresult (*cuDevicePrimaryCtxRetain_t)(CUcontext*, CUdevice);
typedef CUresult (*cuGetProcAddress_t)(const char*, void**, int, cuuint64_t,
                                       CUdriverProcAddressQueryResult*);

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
static cuDevicePrimaryCtxRetain_t real_cuDevicePrimaryCtxRetain;
static cuGetProcAddress_t real_cuGetProcAddress;

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
  real_cuDevicePrimaryCtxRetain =
      (cuDevicePrimaryCtxRetain_t)dlsym(RTLD_NEXT, "cuDevicePrimaryCtxRetain");
  real_cuGetProcAddress =
      (cuGetProcAddress_t)dlsym(RTLD_NEXT, "cuGetProcAddress");
}

static void guard_unsupported(const char* api) {
  if (strstr(api, "nccl") || strstr(api, "NCCL"))
    Tracker::instance().mark_unsupported("nccl");
  if (strstr(api, "Graph") || strstr(api, "graph"))
    Tracker::instance().mark_unsupported("cuda graph");
  if (strstr(api, "MIG") || strstr(api, "mig"))
    Tracker::instance().mark_unsupported("mig");
}

extern "C" {

CUresult cuInit(unsigned int flags) {
  pthread_once(&once, load_syms);
  return real_cuInit(flags);
}

CUresult cuDeviceGet(CUdevice* dev, int ord) {
  pthread_once(&once, load_syms);
  CUresult r = real_cuDeviceGet(dev, ord);
  if (r == CUDA_SUCCESS && dev) Tracker::instance().set_device(*dev);
  return r;
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

CUresult cuLaunchKernel(CUfunction f, unsigned int gx, unsigned int gy,
                        unsigned int gz, unsigned int bx, unsigned int by,
                        unsigned int bz, unsigned int shmem, CUstream st,
                        void** args, void** extra) {
  pthread_once(&once, load_syms);
  guard_unsupported("LaunchKernel");
  return real_cuLaunchKernel(f, gx, gy, gz, bx, by, bz, shmem, st, args, extra);
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
  guard_unsupported(symbol ? symbol : "");
  if (real_cuGetProcAddress)
    return real_cuGetProcAddress(symbol, pfn, cudaVersion, flags, symbolStatus);
  return CUDA_ERROR_NOT_FOUND;
}

}

extern "C" int ckpt_wait_frozen();
extern "C" int ckpt_app_resumed();
