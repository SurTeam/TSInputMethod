cmake_minimum_required(VERSION 3.16)
project(tsime_linux_fcitx5 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(PkgConfig REQUIRED)
pkg_check_modules(Fcitx5Core REQUIRED IMPORTED_TARGET Fcitx5Core)
pkg_check_modules(Fcitx5Module REQUIRED IMPORTED_TARGET Fcitx5Module)

add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../common ${CMAKE_BINARY_DIR}/tsime_common)

add_library(tsime MODULE
  src/tsime_engine.cpp
)
target_include_directories(tsime PRIVATE
  src
  ${CMAKE_CURRENT_SOURCE_DIR}/../common/native
)
target_link_libraries(tsime PRIVATE tsime_ipc PkgConfig::Fcitx5Core PkgConfig::Fcitx5Module)
set_target_properties(tsime PROPERTIES PREFIX "")

install(TARGETS tsime DESTINATION "${FCITX_INSTALL_ADDONDIR}")
install(FILES tsime.conf DESTINATION "${FCITX_INSTALL_PKGDATADIR}/inputmethod")
install(FILES tsime-addon.conf DESTINATION "${FCITX_INSTALL_PKGDATADIR}/addon" RENAME tsime.conf)
