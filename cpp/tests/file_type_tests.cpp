#include "file/file_type.hpp"

#include <array>
#include <cstddef>
#include <cstdint>
#include <cstdlib>
#include <iostream>
#include <stdexcept>
#include <string>
#include <string_view>

namespace {

void require(bool condition, const char* message) {
    if (!condition) {
        throw std::runtime_error(message);
    }
}

cmdr::file::FileInfo makeInfo(std::initializer_list<std::uint8_t> bytes) {
    cmdr::file::FileInfo info{"test.bin", bytes.size()};

    std::size_t index = 0;
    for (const std::uint8_t byte : bytes) {
        info.firstBytes[index++] = byte;
    }
    info.bytesRead = bytes.size();
    return info;
}

void requireType(
    std::initializer_list<std::uint8_t> bytes,
    std::string_view expected
) {
    const cmdr::file::FileInfo info = makeInfo(bytes);
    require(
        cmdr::file::detectFileType(info) == expected,
        ("Unexpected file type: " + std::string(expected)).c_str()
    );
}

} // namespace

int main() {
    try {
        requireType({0x7F, 0x45, 0x4C, 0x46}, "ELF");
        requireType({0x4D, 0x5A}, "DOS/Windows MZ executable");
        requireType({0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "PNG");
        requireType({0xFF, 0xD8, 0xFF}, "JPEG");
        requireType({0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, "GIF87a");
        requireType({0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, "GIF89a");
        requireType({0x25, 0x50, 0x44, 0x46, 0x2D}, "PDF");
        requireType({0x50, 0x4B, 0x03, 0x04}, "ZIP");
        requireType({0x1F, 0x8B}, "GZIP");
        requireType({0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}, "RAR");
        requireType({0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}, "7-Zip");
        requireType({0x42, 0x4D}, "BMP");

        requireType({
            0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00,
            0x57, 0x45, 0x42, 0x50
        }, "WebP");
        requireType({
            0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00,
            0x57, 0x41, 0x56, 0x45
        }, "WAV");
        requireType({0x4F, 0x67, 0x67, 0x53}, "OGG");
        requireType({0x66, 0x4C, 0x61, 0x43}, "FLAC");
        requireType({
            0x53, 0x51, 0x4C, 0x69, 0x74, 0x65,
            0x20, 0x66, 0x6F, 0x72, 0x6D, 0x61,
            0x74, 0x20, 0x33, 0x00
        }, "SQLite");

        requireType({0x52, 0x49, 0x46, 0x46}, "Unknown");
        requireType({
            0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00,
            0x4A, 0x55, 0x4E, 0x4B
        }, "Unknown");
        requireType({
            0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00
        }, "Unknown");
        requireType({
            0x53, 0x51, 0x4C, 0x69, 0x74, 0x65,
            0x20, 0x66, 0x6F, 0x72, 0x6D, 0x61,
            0x74, 0x20, 0x33
        }, "Unknown");

        requireType({0x12, 0x34, 0x56, 0x78}, "Unknown");
        requireType({}, "Unknown");
        requireType({0x7F, 0x45}, "Unknown");

        cmdr::file::FileInfo info{"ignored", 16};
        info.bytesRead = 2;
        info.firstBytes[0] = 0x7F;
        info.firstBytes[1] = 0x45;
        info.firstBytes[2] = 0x4C;
        info.firstBytes[3] = 0x46;
        require(
            cmdr::file::detectFileType(info) == "Unknown",
            "bytesRead must prevent comparison with an uninitialized tail"
        );

        info = cmdr::file::FileInfo{"ignored.wav", 16};
        info.bytesRead = 8;
        info.firstBytes = {
            0x52, 0x49, 0x46, 0x46,
            0x00, 0x00, 0x00, 0x00,
            0x57, 0x41, 0x56, 0x45
        };
        require(
            cmdr::file::detectFileType(info) == "Unknown",
            "bytesRead must prevent reading WAV signatures after the offset limit"
        );

        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "file_type_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
