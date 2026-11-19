cmake_minimum_required(VERSION 3.16)
project(tsime_windows_tsf LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../common ${CMAKE_BINARY_DIR}/tsime_common)

add_library(tsime_tsf SHARED
  src/dllmain.cpp
  src/tsime_tsf.cpp
)
target_include_directories(tsime_tsf PRIVATE
  src
  ${CMAKE_CURRENT_SOURCE_DIR}/../common/native
)
target_link_libraries(tsime_tsf PRIVATE tsime_ipc)
if (WIN32)
  target_link_libraries(tsime_tsf PRIVATE msctf)
endif()
