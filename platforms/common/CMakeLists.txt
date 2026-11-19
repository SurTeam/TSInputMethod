cmake_minimum_required(VERSION 3.16)
project(tsime_ipc LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

find_package(Protobuf REQUIRED)
find_package(PkgConfig REQUIRED)
pkg_check_modules(PROTOBUF REQUIRED protobuf)

set(TSIME_NATIVE_SOURCES
  native/frame.cpp
  native/client.cpp
  native/tsime_c_api.cpp
  generated/ime/v1/ime.pb.cc
)

add_library(tsime_ipc STATIC ${TSIME_NATIVE_SOURCES})
target_include_directories(tsime_ipc
  PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/native
    ${CMAKE_CURRENT_SOURCE_DIR}/generated
    ${PROTOBUF_INCLUDE_DIRS}
)
target_link_libraries(tsime_ipc PUBLIC protobuf::libprotobuf)
target_compile_options(tsime_ipc PRIVATE ${PROTOBUF_CFLAGS_OTHER})

add_executable(tsime_frame_smoke native/frame_smoke.cpp)
target_link_libraries(tsime_frame_smoke PRIVATE tsime_ipc)
