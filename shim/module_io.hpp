#pragma once

int ckpt_modules_write(const char* dir);
int ckpt_modules_restore(const char* dir);
int ckpt_ctx_write(const char* dir);
int ckpt_ctx_restore(const char* dir);
int ckpt_streams_write(const char* dir);
int ckpt_streams_restore(const char* dir);
