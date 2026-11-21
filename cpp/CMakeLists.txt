cmake_minimum_required(VERSION 3.20)

project(cmdr_inspector LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_executable(cmdr-inspector
    src/main.cpp
    src/cli/help.cpp
)

target_include_directories(cmdr-inspector PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/src"
)

set_target_properties(cmdr-inspector PROPERTIES
    CXX_STANDARD 20
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS OFF
)

if(MSVC)
    target_compile_options(cmdr-inspector PRIVATE /W4 /permissive-)
elseif(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang")
    target_compile_options(cmdr-inspector PRIVATE -Wall -Wextra -Wpedantic)
endif()
