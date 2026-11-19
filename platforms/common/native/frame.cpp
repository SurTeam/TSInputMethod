#include "frame.hpp"

#include <cstring>

#ifdef _WIN32
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <windows.h>
#else
#include <unistd.h>
#endif

namespace tsime {
namespace {

bool writeAllFd(int fd, const std::uint8_t *data, std::size_t size) {
  std::size_t offset = 0;
  while (offset < size) {
#ifdef _WIN32
    (void)fd;
    (void)data;
    return false;
#else
    const ssize_t n = ::write(fd, data + offset, size - offset);
    if (n <= 0) {
      return false;
    }
    offset += static_cast<std::size_t>(n);
#endif
  }
  return true;
}

bool readAllFd(int fd, std::uint8_t *data, std::size_t size) {
  std::size_t offset = 0;
  while (offset < size) {
#ifdef _WIN32
    (void)fd;
    (void)data;
    return false;
#else
    const ssize_t n = ::read(fd, data + offset, size - offset);
    if (n <= 0) {
      return false;
    }
    offset += static_cast<std::size_t>(n);
#endif
  }
  return true;
}

void putBe32(std::uint8_t out[4], std::uint32_t value) {
  out[0] = static_cast<std::uint8_t>((value >> 24) & 0xffu);
  out[1] = static_cast<std::uint8_t>((value >> 16) & 0xffu);
  out[2] = static_cast<std::uint8_t>((value >> 8) & 0xffu);
  out[3] = static_cast<std::uint8_t>(value & 0xffu);
}

std::uint32_t getBe32(const std::uint8_t in[4]) {
  return (static_cast<std::uint32_t>(in[0]) << 24) |
         (static_cast<std::uint32_t>(in[1]) << 16) |
         (static_cast<std::uint32_t>(in[2]) << 8) |
         static_cast<std::uint32_t>(in[3]);
}

} // namespace

bool writeFrame(int fd, const std::uint8_t *payload, std::size_t size) {
  if (payload == nullptr || size == 0 || size > kMaxFrameSize) {
    return false;
  }
  std::uint8_t header[4];
  putBe32(header, static_cast<std::uint32_t>(size));
  if (!writeAllFd(fd, header, 4)) {
    return false;
  }
  return writeAllFd(fd, payload, size);
}

bool readFrame(int fd, std::vector<std::uint8_t> *out) {
  if (out == nullptr) {
    return false;
  }
  std::uint8_t header[4];
  if (!readAllFd(fd, header, 4)) {
    return false;
  }
  const std::uint32_t size = getBe32(header);
  if (size == 0 || size > kMaxFrameSize) {
    return false;
  }
  out->assign(size, 0);
  return readAllFd(fd, out->data(), size);
}

#ifdef _WIN32
bool writeFrameHandle(void *handle, const std::uint8_t *payload, std::size_t size) {
  if (handle == nullptr || payload == nullptr || size == 0 || size > kMaxFrameSize) {
    return false;
  }
  std::uint8_t header[4];
  putBe32(header, static_cast<std::uint32_t>(size));
  DWORD written = 0;
  if (!WriteFile(static_cast<HANDLE>(handle), header, 4, &written, nullptr) || written != 4) {
    return false;
  }
  written = 0;
  return WriteFile(static_cast<HANDLE>(handle), payload, static_cast<DWORD>(size), &written, nullptr) &&
         written == size;
}

bool readFrameHandle(void *handle, std::vector<std::uint8_t> *out) {
  if (handle == nullptr || out == nullptr) {
    return false;
  }
  std::uint8_t header[4];
  DWORD readCount = 0;
  if (!ReadFile(static_cast<HANDLE>(handle), header, 4, &readCount, nullptr) || readCount != 4) {
    return false;
  }
  const std::uint32_t size = getBe32(header);
  if (size == 0 || size > kMaxFrameSize) {
    return false;
  }
  out->assign(size, 0);
  readCount = 0;
  return ReadFile(static_cast<HANDLE>(handle), out->data(), static_cast<DWORD>(size), &readCount, nullptr) &&
         readCount == size;
}
#endif

} // namespace tsime
