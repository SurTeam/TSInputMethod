#include "frame.hpp"

#include <cstdio>
#include <string>
#include <vector>

#ifndef _WIN32
#include <sys/socket.h>
#include <unistd.h>
#endif

int main() {
#ifndef _WIN32
  int fds[2];
  if (socketpair(AF_UNIX, SOCK_STREAM, 0, fds) != 0) {
    return 1;
  }
  const std::string payload = "hello-tsime";
  if (!tsime::writeFrame(fds[0], reinterpret_cast<const unsigned char *>(payload.data()), payload.size())) {
    return 2;
  }
  std::vector<unsigned char> out;
  if (!tsime::readFrame(fds[1], &out)) {
    return 3;
  }
  close(fds[0]);
  close(fds[1]);
  if (std::string(out.begin(), out.end()) != payload) {
    return 4;
  }
  std::puts("frame_smoke_ok");
  return 0;
#else
  std::puts("frame_smoke_skipped_windows_host");
  return 0;
#endif
}
