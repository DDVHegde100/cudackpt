#include "quiesce.hpp"
#include "log.h"
#include "tracker.hpp"

int ckpt_quiesce_gpu() {
  auto& tr = Tracker::instance();
  CUresult r = cuCtxSynchronize();
  if (r != CUDA_SUCCESS) return -20;
  for (const auto& sr : tr.streams_snapshot()) {
    r = cuStreamSynchronize(sr.stream);
    if (r != CUDA_SUCCESS) return -21;
  }
  for (const auto& ev : tr.events_snapshot()) {
    r = cuEventSynchronize(ev.event);
    if (r != CUDA_SUCCESS) return -22;
  }
  int spins = 0;
  while (tr.pending_host_callbacks() > 0 && spins < 100000) {
    r = cuCtxSynchronize();
    if (r != CUDA_SUCCESS) return -23;
    ++spins;
  }
  if (tr.pending_host_callbacks() > 0) {
    ckpt_log("cudackpt: quiesce timeout callbacks=%zu\n", tr.pending_host_callbacks());
    return -24;
  }
  return 0;
}
