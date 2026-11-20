#include "tsime_c_api.h"

#include "client.hpp"

#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>

struct TsimeClient {
  tsime::Client inner;
};

namespace {

char *copyCString(const std::string &value) {
  char *out = static_cast<char *>(std::malloc(value.size() + 1));
  if (out == nullptr) {
    return nullptr;
  }
  std::memcpy(out, value.c_str(), value.size() + 1);
  return out;
}

TsimeKeyResult toCResult(const tsime::KeyOutcome &outcome) {
  TsimeKeyResult result;
  std::memset(&result, 0, sizeof(result));
  result.ok = outcome.ok ? 1 : 0;
  result.consumed = outcome.consumed ? 1 : 0;
  result.session_id = copyCString(outcome.sessionId);
  result.preedit = copyCString(outcome.preedit);
  result.cursor = outcome.cursor;
  result.generation = outcome.generation;
  result.commit_text = copyCString(outcome.commitText);
  result.selected_index = outcome.selectedIndex;
  result.error = copyCString(outcome.error);
  if (!outcome.candidates.empty()) {
    auto *items = static_cast<TsimeCandidate *>(std::calloc(outcome.candidates.size(), sizeof(TsimeCandidate)));
    if (items != nullptr) {
      for (size_t i = 0; i < outcome.candidates.size(); ++i) {
        items[i].text = copyCString(outcome.candidates[i]);
        items[i].comment = copyCString(i < outcome.comments.size() ? outcome.comments[i] : std::string());
        items[i].score = i < outcome.scores.size() ? outcome.scores[i] : 0;
      }
      result.candidates = items;
      result.candidate_count = outcome.candidates.size();
    }
  }
  return result;
}

} // namespace

extern "C" TsimeClient *tsime_client_create(void) { return new TsimeClient(); }

extern "C" void tsime_client_destroy(TsimeClient *client) { delete client; }

extern "C" int tsime_client_connect(TsimeClient *client, const char *path) {
  if (client == nullptr || path == nullptr) {
    return 0;
  }
  return client->inner.connectPath(path) ? 1 : 0;
}

extern "C" void tsime_client_disconnect(TsimeClient *client) {
  if (client != nullptr) {
    client->inner.disconnect();
  }
}

extern "C" int tsime_client_hello(TsimeClient *client, const char *client_name) {
  if (client == nullptr) {
    return 0;
  }
  return client->inner.hello(client_name == nullptr ? "adapter" : client_name) ? 1 : 0;
}

extern "C" int tsime_client_open_session(TsimeClient *client, int mode, char *session_id_out, size_t session_id_cap) {
  if (client == nullptr || session_id_out == nullptr || session_id_cap == 0) {
    return 0;
  }
  std::string sessionId;
  if (!client->inner.openSession(mode, &sessionId)) {
    return 0;
  }
  if (sessionId.size() + 1 > session_id_cap) {
    return 0;
  }
  std::memcpy(session_id_out, sessionId.c_str(), sessionId.size() + 1);
  return 1;
}

extern "C" int tsime_client_close_session(TsimeClient *client, const char *session_id) {
  if (client == nullptr || session_id == nullptr) {
    return 0;
  }
  return client->inner.closeSession(session_id) ? 1 : 0;
}

extern "C" int tsime_client_invalidate_learning_context(TsimeClient *client, const char *session_id) {
  if (client == nullptr || session_id == nullptr) {
    return 0;
  }
  return client->inner.invalidateLearningContext(session_id) ? 1 : 0;
}

extern "C" int tsime_client_set_mode(TsimeClient *client, const char *session_id, int mode) {
  if (client == nullptr || session_id == nullptr) {
    return 0;
  }
  return client->inner.setMode(session_id, mode) ? 1 : 0;
}

extern "C" TsimeKeyResult tsime_client_handle_key(TsimeClient *client, const char *session_id, uint32_t key_code,
                                                  uint32_t modifiers, const char *characters, int key_down,
                                                  int is_repeat, uint64_t generation) {
  if (client == nullptr || session_id == nullptr) {
    TsimeKeyResult result;
    std::memset(&result, 0, sizeof(result));
    result.error = copyCString("invalid client");
    return result;
  }
  const auto outcome = client->inner.handleKey(session_id, key_code, modifiers, characters == nullptr ? "" : characters,
                                               key_down != 0, is_repeat != 0, generation);
  return toCResult(outcome);
}

extern "C" TsimeKeyResult tsime_client_select_candidate(TsimeClient *client, const char *session_id,
                                                        uint64_t generation, uint32_t index) {
  if (client == nullptr || session_id == nullptr) {
    TsimeKeyResult result;
    std::memset(&result, 0, sizeof(result));
    result.error = copyCString("invalid client");
    return result;
  }
  return toCResult(client->inner.selectCandidate(session_id, generation, index));
}

extern "C" void tsime_key_result_free(TsimeKeyResult *result) {
  if (result == nullptr) {
    return;
  }
  std::free(const_cast<char *>(result->session_id));
  std::free(const_cast<char *>(result->preedit));
  std::free(const_cast<char *>(result->commit_text));
  std::free(const_cast<char *>(result->error));
  if (result->candidates != nullptr) {
    for (size_t i = 0; i < result->candidate_count; ++i) {
      std::free(const_cast<char *>(result->candidates[i].text));
      std::free(const_cast<char *>(result->candidates[i].comment));
    }
    std::free(const_cast<TsimeCandidate *>(result->candidates));
  }
  std::memset(result, 0, sizeof(*result));
}
