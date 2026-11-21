#include "file/file_type.hpp"

#include <array>
#include <cstddef>
#include <cstdint>
#include <initializer_list>

namespace cmdr::file {

namespace {

bool matchesAt(
    const FileInfo& info,
    std::size_t offset,
    std::initializer_list<std::uint8_t> signature
) {
    if (info.bytesRead < offset + signature.size()) {
        return false;
    }

    std::size_t index = 0;
    for (const std::uint8_t byte : signature) {
        if (info.firstBytes[offset + index] != byte) {
            return false;
        }
        ++index;
    }

    return true;
}

bool startsWith(
    const FileInfo& info,
    std::initializer_list<std::uint8_t> signature
) {
    return matchesAt(info, 0, signature);
}

} // namespace

std::string_view detectFileType(const FileInfo& info) {
    if (startsWith(info, {0x7F, 0x45, 0x4C, 0x46})) return "ELF";
    if (startsWith(info, {0x4D, 0x5A})) return "DOS/Windows MZ executable";

    if (startsWith(info, {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})) return "PNG";
    if (startsWith(info, {0xFF, 0xD8, 0xFF})) return "JPEG";

    if (startsWith(info, {0x47, 0x49, 0x46, 0x38, 0x37, 0x61})) return "GIF87a";
    if (startsWith(info, {0x47, 0x49, 0x46, 0x38, 0x39, 0x61})) return "GIF89a";

    if (startsWith(info, {0x25, 0x50, 0x44, 0x46, 0x2D})) return "PDF";
    if (startsWith(info, {0x50, 0x4B, 0x03, 0x04})) return "ZIP";
    if (startsWith(info, {0x1F, 0x8B})) return "GZIP";

    if (startsWith(info, {0x52, 0x61, 0x72, 0x21, 0x1A, 0x07})) return "RAR";
    if (startsWith(info, {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C})) return "7-Zip";
    if (startsWith(info, {0x42, 0x4D})) return "BMP";

    if (startsWith(info, {0x4F, 0x67, 0x67, 0x53})) return "OGG";
    if (startsWith(info, {0x66, 0x4C, 0x61, 0x43})) return "FLAC";

    if (startsWith(info, {
            0x53, 0x51, 0x4C, 0x69, 0x74, 0x65,
            0x20, 0x66, 0x6F, 0x72, 0x6D, 0x61,
            0x74, 0x20, 0x33, 0x00
        })) return "SQLite";

    if (matchesAt(info, 0, {0x52, 0x49, 0x46, 0x46}) &&
        matchesAt(info, 8, {0x57, 0x45, 0x42, 0x50})) return "WebP";

    if (matchesAt(info, 0, {0x52, 0x49, 0x46, 0x46}) &&
        matchesAt(info, 8, {0x57, 0x41, 0x56, 0x45})) return "WAV";

    return "Unknown";
}

} // namespace cmdr::file
