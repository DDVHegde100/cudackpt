#include <cublas_v2.h>
#include <cuda_runtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

extern "C" int ckpt_wait_frozen();
extern "C" int ckpt_app_resumed();

static void wr(const char* p, const char* s) {
  FILE* f = fopen(p, "w");
  if (f) {
    fputs(s, f);
    fclose(f);
  }
}

int main() {
  const char* ready = "/tmp/cublas.ready";
  const char* out = "/tmp/cublas.out";
  unlink(ready);
  unlink(out);
  cublasHandle_t handle;
  if (cublasCreate(&handle) != CUBLAS_STATUS_SUCCESS) return 1;
  const int m = 128, n = 128, k = 128;
  float *d_a, *d_b, *d_c;
  size_t sa = static_cast<size_t>(m * k) * sizeof(float);
  size_t sb = static_cast<size_t>(k * n) * sizeof(float);
  size_t sc = static_cast<size_t>(m * n) * sizeof(float);
  if (cudaMalloc(&d_a, sa) != cudaSuccess) return 1;
  if (cudaMalloc(&d_b, sb) != cudaSuccess) return 1;
  if (cudaMalloc(&d_c, sc) != cudaSuccess) return 1;
  float* h_a = (float*)malloc(sa);
  float* h_b = (float*)malloc(sb);
  float* h_c = (float*)malloc(sc);
  for (int i = 0; i < m * k; ++i) h_a[i] = 1.0f;
  for (int i = 0; i < k * n; ++i) h_b[i] = 2.0f;
  cudaMemcpy(d_a, h_a, sa, cudaMemcpyHostToDevice);
  cudaMemcpy(d_b, h_b, sb, cudaMemcpyHostToDevice);
  float alpha = 1.0f, beta = 0.0f;
  const int loops = 50;
  const int ckpt_at = loops / 2;
  for (int i = 0; i < loops; ++i) {
    if (cublasSgemm(handle, CUBLAS_OP_N, CUBLAS_OP_N, m, n, k, &alpha, d_a, m, d_b, k, &beta, d_c,
                    m) != CUBLAS_STATUS_SUCCESS) {
      return 1;
    }
    cudaDeviceSynchronize();
    if (i == ckpt_at) {
      wr(ready, "1");
      ckpt_wait_frozen();
      ckpt_app_resumed();
    }
  }
  cudaMemcpy(h_c, d_c, sc, cudaMemcpyDeviceToHost);
  double sum = 0;
  for (int i = 0; i < m * n; ++i) sum += h_c[i];
  char buf[64];
  snprintf(buf, sizeof(buf), "%.6f\n", sum);
  wr(out, buf);
  cublasDestroy(handle);
  cudaFree(d_a);
  cudaFree(d_b);
  cudaFree(d_c);
  free(h_a);
  free(h_b);
  free(h_c);
  return 0;
}
