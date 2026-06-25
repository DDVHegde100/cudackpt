#include "coalesce.hpp"
#include <algorithm>

std::vector<CoalescedAlloc> coalesce_allocs(const std::vector<AllocRec>& allocs) {
  if (allocs.empty()) return {};
  std::vector<CoalescedAlloc> out;
  out.reserve(allocs.size());
  CoalescedAlloc cur{};
  cur.base = allocs[0].base;
  cur.size = allocs[0].size;
  cur.ctx = allocs[0].ctx;
  cur.flags = allocs[0].flags;
  cur.seq = allocs[0].seq;
  cur.parts.push_back(allocs[0]);
  for (size_t i = 1; i < allocs.size(); ++i) {
    const auto& a = allocs[i];
    uint64_t end = static_cast<uint64_t>(cur.base) + cur.size;
    if (a.ctx == cur.ctx && static_cast<uint64_t>(a.base) == end) {
      cur.size += a.size;
      cur.parts.push_back(a);
      continue;
    }
    out.push_back(cur);
    cur = CoalescedAlloc{};
    cur.base = a.base;
    cur.size = a.size;
    cur.ctx = a.ctx;
    cur.flags = a.flags;
    cur.seq = a.seq;
    cur.parts = {a};
  }
  out.push_back(cur);
  return out;
}
