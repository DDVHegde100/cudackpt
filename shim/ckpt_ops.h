#pragma once
#include <stdint.h>
#include <stddef.h>

typedef struct {
  uint64_t ptr;
  uint64_t size;
  uint64_t offset;
  uint32_t crc32c;
  uint32_t pad;
} ChunkEntry;

typedef struct {
  uint32_t magic;
  uint16_t version;
  uint16_t flags;
  uint32_t count;
  uint64_t total_bytes;
} ManifestHeader;

int ckpt_snapshot_write(const char* dir, ChunkEntry* entries, int* count);
int ckpt_restore_load(const char* dir, ChunkEntry* entries, int max, int* count);
int ckpt_resume_signal();
