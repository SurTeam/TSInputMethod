#include "client.hpp"

#include "frame.hpp"

#include <cstring>

#ifdef _WIN32
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <windows.h>
#else
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>
#endif

namespace tsime {

Client::~Client() { disconnect(); }

bool Client::connectPath(const std::string &path) {
  std::lock_guard<std::mutex> lock(mu_);
  disconnect();
#ifdef _WIN32
  HANDLE handle = CreateFileA(path.c_str(), GENERIC_READ | GENERIC_WRITE, 0, nullptr, OPEN_EXISTING, 0, nullptr);
  if (handle == INVALID_HANDLE_VALUE) {
    return false;
  }
  pipe_ = handle;
  return true;
#else
  const int fd = ::socket(AF_UNIX, SOCK_STREAM, 0);
  if (fd < 0) {
    return false;
  }
  sockaddr_un addr;
  std::memset(&addr, 0, sizeof(addr));
  addr.sun_family = AF_UNIX;
  if (path.size() >= sizeof(addr.sun_path)) {
    ::close(fd);
    return false;
  }
  std::memcpy(addr.sun_path, path.c_str(), path.size() + 1);
  if (::connect(fd, reinterpret_cast<sockaddr *>(&addr), sizeof(addr)) != 0) {
    ::close(fd);
    return false;
  }
  fd_ = fd;
  return true;
#endif
}

void Client::disconnect() {
#ifdef _WIN32
  if (pipe_ != nullptr) {
    CloseHandle(static_cast<HANDLE>(pipe_));
    pipe_ = nullptr;
  }
#endif
  if (fd_ >= 0) {
#ifndef _WIN32
    ::close(fd_);
#endif
    fd_ = -1;
  }
}

bool Client::connected() const {
  std::lock_guard<std::mutex> lock(mu_);
#ifdef _WIN32
  return pipe_ != nullptr;
#else
  return fd_ >= 0;
#endif
}

bool Client::writePayload(const std::string &bytes) {
#ifdef _WIN32
  return writeFrameHandle(pipe_, reinterpret_cast<const std::uint8_t *>(bytes.data()), bytes.size());
#else
  return writeFrame(fd_, reinterpret_cast<const std::uint8_t *>(bytes.data()), bytes.size());
#endif
}

bool Client::readPayload(std::string *bytes) {
  std::vector<std::uint8_t> payload;
#ifdef _WIN32
  if (!readFrameHandle(pipe_, &payload)) {
    return false;
  }
#else
  if (!readFrame(fd_, &payload)) {
    return false;
  }
#endif
  bytes->assign(reinterpret_cast<const char *>(payload.data()), payload.size());
  return true;
}

bool Client::roundTrip(const ime::v1::Envelope &req, ime::v1::Envelope *resp) {
  std::string encoded;
  if (!req.SerializeToString(&encoded)) {
    return false;
  }
  if (!writePayload(encoded)) {
    return false;
  }
  std::string raw;
  if (!readPayload(&raw)) {
    return false;
  }
  return resp->ParseFromString(raw);
}

bool Client::hello(const std::string &clientName) {
  std::lock_guard<std::mutex> lock(mu_);
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  auto *hello = req.mutable_hello();
  hello->mutable_client_version()->set_major(1);
  hello->mutable_client_version()->set_minor(0);
  hello->set_client_name(clientName);
  ime::v1::Envelope resp;
  if (!roundTrip(req, &resp)) {
    return false;
  }
  if (!resp.has_hello_ack()) {
    return false;
  }
  return resp.hello_ack().accepted();
}

bool Client::openSession(int mode, std::string *sessionIdOut) {
  std::lock_guard<std::mutex> lock(mu_);
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  auto *open = req.mutable_open_session();
  open->set_initial_mode(static_cast<ime::v1::InputMode>(mode));
  ime::v1::Envelope resp;
  if (!roundTrip(req, &resp) || !resp.has_open_session_ack()) {
    return false;
  }
  if (sessionIdOut != nullptr) {
    *sessionIdOut = resp.open_session_ack().session_id();
  }
  return !resp.open_session_ack().session_id().empty();
}

bool Client::closeSession(const std::string &sessionId) {
  std::lock_guard<std::mutex> lock(mu_);
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  req.mutable_close_session()->set_session_id(sessionId);
  ime::v1::Envelope resp;
  return roundTrip(req, &resp);
}

bool Client::invalidateLearningContext(const std::string &sessionId) {
  std::lock_guard<std::mutex> lock(mu_);
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  req.mutable_invalidate_learning_context()->set_session_id(sessionId);
  ime::v1::Envelope resp;
  return roundTrip(req, &resp) && resp.has_status() && resp.status().ok();
}

bool Client::setMode(const std::string &sessionId, int mode) {
  std::lock_guard<std::mutex> lock(mu_);
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  auto *setMode = req.mutable_set_input_mode();
  setMode->set_session_id(sessionId);
  setMode->set_mode(static_cast<ime::v1::InputMode>(mode));
  ime::v1::Envelope resp;
  return roundTrip(req, &resp);
}

KeyOutcome Client::handleKey(const std::string &sessionId, std::uint32_t keyCode, std::uint32_t modifiers,
                             const std::string &characters, bool keyDown, bool isRepeat,
                             std::uint64_t generation) {
  std::lock_guard<std::mutex> lock(mu_);
  KeyOutcome out;
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  req.set_generation(generation);
  auto *key = req.mutable_key_event();
  key->set_session_id(sessionId);
  key->set_key_code(keyCode);
  key->set_modifiers(modifiers);
  key->set_characters(characters);
  key->set_key_down(keyDown);
  key->set_is_repeat(isRepeat);
  ime::v1::Envelope resp;
  if (!roundTrip(req, &resp)) {
    out.error = "roundtrip failed";
    return out;
  }
  if (resp.has_error()) {
    out.error = resp.error().message();
    return out;
  }
  if (!resp.has_key_event_result()) {
    out.error = "missing key_event_result";
    return out;
  }
  const auto &result = resp.key_event_result();
  out.ok = true;
  out.consumed = result.consumed();
  if (result.has_composition()) {
    out.sessionId = result.composition().session_id();
    out.preedit = result.composition().preedit();
    out.cursor = result.composition().cursor();
    out.generation = result.composition().generation();
  }
  if (result.has_commit()) {
    out.commitText = result.commit().text();
    out.generation = result.commit().generation();
    out.sessionId = result.commit().session_id();
  }
  if (result.has_candidates()) {
    out.generation = result.candidates().generation();
    out.selectedIndex = result.candidates().selected_index();
    for (const auto &cand : result.candidates().candidates()) {
      out.candidates.push_back(cand.text());
      out.comments.push_back(cand.comment());
      out.scores.push_back(cand.score());
    }
  }
  return out;
}

KeyOutcome Client::selectCandidate(const std::string &sessionId, std::uint64_t generation, std::uint32_t index) {
  std::lock_guard<std::mutex> lock(mu_);
  KeyOutcome out;
  ime::v1::Envelope req;
  req.set_request_id(nextRequestId_++);
  req.set_sequence_number(nextSequence_++);
  req.set_generation(generation);
  auto *select = req.mutable_select_candidate();
  select->set_session_id(sessionId);
  select->set_generation(generation);
  select->set_index(index);
  ime::v1::Envelope resp;
  if (!roundTrip(req, &resp)) {
    out.error = "roundtrip failed";
    return out;
  }
  if (resp.has_error()) {
    out.error = resp.error().message();
    return out;
  }
  if (resp.has_key_event_result()) {
    const auto &result = resp.key_event_result();
    out.ok = true;
    out.consumed = result.consumed();
    if (result.has_commit()) {
      out.commitText = result.commit().text();
      out.generation = result.commit().generation();
      out.sessionId = result.commit().session_id();
    }
    if (result.has_composition()) {
      out.preedit = result.composition().preedit();
      out.cursor = result.composition().cursor();
      out.generation = result.composition().generation();
      out.sessionId = result.composition().session_id();
    }
    if (result.has_candidates()) {
      out.generation = result.candidates().generation();
      out.selectedIndex = result.candidates().selected_index();
      for (const auto &cand : result.candidates().candidates()) {
        out.candidates.push_back(cand.text());
        out.comments.push_back(cand.comment());
        out.scores.push_back(cand.score());
      }
    }
  } else if (resp.has_commit_text()) {
    out.ok = true;
    out.consumed = true;
    out.commitText = resp.commit_text().text();
    out.generation = resp.commit_text().generation();
    out.sessionId = resp.commit_text().session_id();
  } else {
    out.error = "unexpected select response";
  }
  return out;
}

} // namespace tsime
