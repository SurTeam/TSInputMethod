#pragma once

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct TsimeClient TsimeClient;

typedef struct TsimeCandidate {
  const char *text;
  const char *comment;
  int32_t score;
} TsimeCandidate;

typedef struct TsimeKeyResult {
  int ok;
  int consumed;
  const char *session_id;
  const char *preedit;
  uint32_t cursor;
  uint64_t generation;
  const char *commit_text;
  const TsimeCandidate *candidates;
  size_t candidate_count;
  uint32_t selected_index;
  const char *error;
} TsimeKeyResult;

TsimeClient *tsime_client_create(void);
void tsime_client_destroy(TsimeClient *client);
int tsime_client_connect(TsimeClient *client, const char *path);
void tsime_client_disconnect(TsimeClient *client);
int tsime_client_hello(TsimeClient *client, const char *client_name);
int tsime_client_open_session(TsimeClient *client, int mode, char *session_id_out, size_t session_id_cap);
int tsime_client_close_session(TsimeClient *client, const char *session_id);
int tsime_client_set_mode(TsimeClient *client, const char *session_id, int mode);
TsimeKeyResult tsime_client_handle_key(TsimeClient *client, const char *session_id, uint32_t key_code,
                                       uint32_t modifiers, const char *characters, int key_down, int is_repeat,
                                       uint64_t generation);
TsimeKeyResult tsime_client_select_candidate(TsimeClient *client, const char *session_id, uint64_t generation,
                                             uint32_t index);
void tsime_key_result_free(TsimeKeyResult *result);

enum {
  TSIME_MODE_UNSPECIFIED = 0,
  TSIME_MODE_CHINESE = 1,
  TSIME_MODE_ENGLISH = 2
};

#ifdef __cplusplus
}
#endif
