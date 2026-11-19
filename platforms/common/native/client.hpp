#pragma once

#include <cstdint>
#include <mutex>
#include <string>
#include <vector>

#include "ime/v1/ime.pb.h"

namespace tsime {

struct KeyOutcome {
  bool ok = false;
  bool consumed = false;
  std::string sessionId;
  std::string preedit;
  std::uint32_t cursor = 0;
  std::uint64_t generation = 0;
  std::string commitText;
  std::vector<std::string> candidates;
  std::vector<std::string> comments;
  std::vector<std::int32_t> scores;
  std::string error;
};

class Client {
public:
  Client() = default;
  ~Client();

  Client(const Client &) = delete;
  Client &operator=(const Client &) = delete;

  bool connectPath(const std::string &path);
  void disconnect();
  bool connected() const;

  bool hello(const std::string &clientName);
  bool openSession(int mode, std::string *sessionIdOut);
  bool closeSession(const std::string &sessionId);
  bool setMode(const std::string &sessionId, int mode);
  KeyOutcome handleKey(const std::string &sessionId, std::uint32_t keyCode, std::uint32_t modifiers,
                       const std::string &characters, bool keyDown, bool isRepeat,
                       std::uint64_t generation);
  KeyOutcome selectCandidate(const std::string &sessionId, std::uint64_t generation, std::uint32_t index);

private:
  bool roundTrip(const ime::v1::Envelope &req, ime::v1::Envelope *resp);
  bool writePayload(const std::string &bytes);
  bool readPayload(std::string *bytes);

  mutable std::mutex mu_;
  int fd_ = -1;
#ifdef _WIN32
  void *pipe_ = nullptr;
#endif
  std::uint64_t nextRequestId_ = 1;
  std::uint64_t nextSequence_ = 1;
};

}