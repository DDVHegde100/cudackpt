#include "ckpt_ops.h"
#include "tracker.hpp"
#include "log.h"
#include <arpa/inet.h>
#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <pthread.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/un.h>
#include <unistd.h>

#ifndef SIGCKPT
#define SIGCKPT SIGUSR2
#endif

enum {
  OP_PING = 1,
  OP_STATUS = 2,
  OP_FREEZE = 3,
  OP_SNAPSHOT = 4,
  OP_RESTORE = 5,
  OP_RESUME = 6,
  OP_QUIT = 7
};

enum {
  ST_IDLE = 0,
  ST_FROZEN = 1,
  ST_SNAPPED = 2,
  ST_RESTORED = 3,
  ST_RUNNING = 4
};

static int g_sock = -1;
static int g_state = ST_IDLE;
static pthread_mutex_t g_mu = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t g_cv = PTHREAD_COND_INITIALIZER;
static char g_image_dir[512];
static volatile sig_atomic_t g_sigckpt = 0;

extern int ckpt_snapshot_write(const char* dir, ChunkEntry* entries, int* count);
extern int ckpt_restore_load(const char* dir, ChunkEntry* entries, int max, int* count);
extern int ckpt_resume_signal();

static void write_u32(int fd, uint32_t v) {
  uint32_t be = htonl(v);
  write(fd, &be, 4);
}

static uint32_t read_u32(int fd) {
  uint32_t be = 0;
  if (read(fd, &be, 4) != 4) return 0;
  return ntohl(be);
}

static void write_str(int fd, const char* s) {
  uint32_t n = s ? (uint32_t)strlen(s) : 0;
  write_u32(fd, n);
  if (n) write(fd, s, n);
}

static int read_str(int fd, char* buf, size_t cap) {
  uint32_t n = read_u32(fd);
  if (n >= cap) return -1;
  if (n == 0) {
    buf[0] = 0;
    return 0;
  }
  if (read(fd, buf, n) != (ssize_t)n) return -1;
  buf[n] = 0;
  return 0;
}

static void set_state(int s) {
  pthread_mutex_lock(&g_mu);
  g_state = s;
  Tracker::instance().set_state(s);
  pthread_cond_broadcast(&g_cv);
  pthread_mutex_unlock(&g_mu);
}

static void handle_client(int cfd) {
  uint32_t op = read_u32(cfd);
  char path[512];
  switch (op) {
    case OP_PING:
      write_u32(cfd, 0);
      break;
    case OP_STATUS:
      pthread_mutex_lock(&g_mu);
      write_u32(cfd, (uint32_t)g_state);
      pthread_mutex_unlock(&g_mu);
      break;
    case OP_FREEZE:
      set_state(ST_FROZEN);
      write_u32(cfd, 0);
      break;
    case OP_SNAPSHOT:
      if (read_str(cfd, path, sizeof(path)) != 0) {
        write_u32(cfd, 1);
        break;
      }
      strncpy(g_image_dir, path, sizeof(g_image_dir) - 1);
      Tracker::instance().set_image_dir(path);
      {
        int cnt = 0;
        int rc = ckpt_snapshot_write(path, NULL, &cnt);
        if (rc == 0) set_state(ST_SNAPPED);
        write_u32(cfd, rc == 0 ? 0u : 1u);
      }
      break;
    case OP_RESTORE:
      if (read_str(cfd, path, sizeof(path)) != 0) {
        write_u32(cfd, 1);
        break;
      }
      {
        int cnt = 0;
        int rc = ckpt_restore_load(path, NULL, 0, &cnt);
        if (rc == 0) set_state(ST_RESTORED);
        write_u32(cfd, rc == 0 ? 0u : 1u);
      }
      break;
    case OP_RESUME:
      ckpt_resume_signal();
      set_state(ST_RUNNING);
      write_u32(cfd, 0);
      break;
    case OP_QUIT:
      write_u32(cfd, 0);
      close(cfd);
      return;
    default:
      write_u32(cfd, 1);
      break;
  }
  close(cfd);
}

static void* server_thread(void* arg) {
  (void)arg;
  for (;;) {
    int cfd = accept(g_sock, NULL, NULL);
    if (cfd < 0) continue;
    handle_client(cfd);
  }
  return NULL;
}

static void sig_handler(int sig) {
  (void)sig;
  g_sigckpt = 1;
  set_state(ST_FROZEN);
}

static void ensure_run_dir() {
  mkdir("/run/cudackpt", 0755);
}

__attribute__((constructor)) static void ckpt_ipc_init() {
  ensure_run_dir();
  char path[256];
  snprintf(path, sizeof(path), "/run/cudackpt/%d.sock", getpid());
  unlink(path);
  g_sock = socket(AF_UNIX, SOCK_STREAM, 0);
  if (g_sock < 0) return;
  struct sockaddr_un addr;
  memset(&addr, 0, sizeof(addr));
  addr.sun_family = AF_UNIX;
  strncpy(addr.sun_path, path, sizeof(addr.sun_path) - 1);
  if (bind(g_sock, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
    close(g_sock);
    g_sock = -1;
    return;
  }
  listen(g_sock, 8);
  pthread_t tid;
  pthread_create(&tid, NULL, server_thread, NULL);
  pthread_detach(tid);
  struct sigaction sa;
  memset(&sa, 0, sizeof(sa));
  sa.sa_handler = sig_handler;
  sigemptyset(&sa.sa_mask);
  sigaction(SIGCKPT, &sa, NULL);
  if (getenv("CUDACKPT_NEED_GPU"))
    set_state(ST_FROZEN);
  else
    set_state(ST_RUNNING);
}

int ckpt_wait_frozen() {
  pthread_mutex_lock(&g_mu);
  while (g_state != ST_FROZEN && g_state != ST_SNAPPED) pthread_cond_wait(&g_cv, &g_mu);
  int s = g_state;
  pthread_mutex_unlock(&g_mu);
  return s;
}

int ckpt_app_resumed() {
  pthread_mutex_lock(&g_mu);
  while (g_state != ST_RUNNING && g_state != ST_RESTORED) pthread_cond_wait(&g_cv, &g_mu);
  int s = g_state;
  pthread_mutex_unlock(&g_mu);
  return s;
}
