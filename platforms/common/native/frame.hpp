#pragma once

#include <cstddef>
#include <cstdint>
#include <string>
#include <vector>

namespace tsime {

constexpr std::size_t kMaxFrameSize = 4u * 1024u * 1024u;

bool writeFrame(int fd, const std::uint8_t *payload, std::size_t size);
bool readFrame(int fd, std::vector<std::uint8_t> *out);

#ifdef _WIN32
bool writeFrameHandle(void *handle, const std::uint8_t *payload, std::size_t size);
bool readFrameHandle(void *handle, std::vector<std::uint8_t> *out);
#endif

}