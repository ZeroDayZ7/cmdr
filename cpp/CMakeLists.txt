cmake_minimum_required(VERSION 3.20)

project(cmdr_inspector LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_executable(cmdr-inspector
    src/main.cpp
    src/cli/help.cpp
    src/file/file_inspector.cpp
    src/file/file_type.cpp
    src/file/extension_check.cpp
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

include(CTest)

if(BUILD_TESTING)
    add_executable(file_type_tests
        tests/file_type_tests.cpp
        src/file/file_type.cpp
    )
    target_include_directories(file_type_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/src"
    )

    add_executable(file_inspector_tests
        tests/file_inspector_tests.cpp
        src/file/file_inspector.cpp
    )
    target_include_directories(file_inspector_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/src"
    )

    add_executable(extension_check_tests
        tests/extension_check_tests.cpp
        src/file/extension_check.cpp
    )
    target_include_directories(extension_check_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/src"
    )

    add_executable(cli_tests
        tests/cli_tests.cpp
    )
    target_include_directories(cli_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/src"
    )

    add_test(NAME file_type_tests COMMAND file_type_tests)
    add_test(NAME file_inspector_tests COMMAND file_inspector_tests)
    add_test(NAME extension_check_tests COMMAND extension_check_tests)
    add_test(NAME cli_tests COMMAND cli_tests
        "$<TARGET_FILE:cmdr-inspector>"
        "${CMAKE_CURRENT_SOURCE_DIR}/test-data/samples"
    )
endif()
