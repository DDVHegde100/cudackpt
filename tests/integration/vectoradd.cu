#include <cuda_runtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

extern "C" int ckpt_wait_frozen();
extern "C" int ckpt_app_resumed();

__global__ void vadd(const float* a, const float* b, float* c, int n) {
  int i = blockIdx.x * blockDim.x + threadIdx.x;
  if (i < n) c[i] = a[i] + b[i];
}

static void wr(const char* p, const char* s) {
  FILE* f = fopen(p, "w");
  if (f) {
    fputs(s, f);
    fclose(f);
  }
}

int main() {
  const char* ready = "/tmp/vectoradd.ready";
  const char* out = "/tmp/vectoradd.out";
  unlink(ready);
  unlink(out);
  const int n = 1 << 20;
  float *d_a, *d_b, *d_c;
  size_t bytes = static_cast<size_t>(n) * sizeof(float);
  if (cudaMalloc(&d_a, bytes) != cudaSuccess) return 1;
  if (cudaMalloc(&d_b, bytes) != cudaSuccess) return 1;
  if (cudaMalloc(&d_c, bytes) != cudaSuccess) return 1;
  float *h_a = (float*)malloc(bytes);
  float *h_b = (float*)malloc(bytes);
  for (int i = 0; i < n; ++i) {
    h_a[i] = 1.0f;
    h_b[i] = 2.0f;
  }
  cudaMemcpy(d_a, h_a, bytes, cudaMemcpyHostToDevice);
  cudaMemcpy(d_b, h_b, bytes, cudaMemcpyHostToDevice);
  const int loops = 100;
  const int ckpt_at = loops / 2;
  int bx = 256;
  int gx = (n + bx - 1) / bx;
  for (int i = 0; i < loops; ++i) {
    vadd<<<gx, bx>>>(d_a, d_b, d_c, n);
    cudaDeviceSynchronize();
    if (i == ckpt_at) {
      wr(ready, "1");
      ckpt_wait_frozen();
      ckpt_app_resumed();
    }
  }
  float* h_c = (float*)malloc(bytes);
  cudaMemcpy(h_c, d_c, bytes, cudaMemcpyDeviceToHost);
  double sum = 0;
  for (int i = 0; i < n; ++i) sum += h_c[i];
  char buf[64];
  snprintf(buf, sizeof(buf), "%.6f\n", sum);
  wr(out, buf);
  cudaFree(d_a);
  cudaFree(d_b);
  cudaFree(d_c);
  free(h_a);
  free(h_b);
  free(h_c);
  return 0;
}
