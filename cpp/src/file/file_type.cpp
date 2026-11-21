#include "file/file_type.hpp"

#include <array>
#include <cstddef>
#include <cstdint>

namespace cmdr::file {

namespace {

bool startsWith(
    const FileInfo& info,
    const std::array<std::uint8_t, 16>& signature,
    std::size_t signatureSize
) {
    if (info.bytesRead < signatureSize) {
        return false;
    }

    for (std::size_t index = 0; index < signatureSize; ++index) {
        if (info.firstBytes[index] != signature[index]) {
            return false;
        }
    }

    return true;
}

} // namespace

std::string_view detectFileType(const FileInfo& info) {
    std::array<std::uint8_t, 16> signature{};

    signature = {0x7F, 0x45, 0x4C, 0x46};
    if (startsWith(info, signature, 4)) return "ELF";

    signature = {0x4D, 0x5A};
    if (startsWith(info, signature, 2)) return "DOS/Windows MZ executable";

    signature = {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A};
    if (startsWith(info, signature, 8)) return "PNG";

    signature = {0xFF, 0xD8, 0xFF};
    if (startsWith(info, signature, 3)) return "JPEG";

    signature = {0x47, 0x49, 0x46, 0x38, 0x37, 0x61};
    if (startsWith(info, signature, 6)) return "GIF87a";

    signature = {0x47, 0x49, 0x46, 0x38, 0x39, 0x61};
    if (startsWith(info, signature, 6)) return "GIF89a";

    signature = {0x25, 0x50, 0x44, 0x46, 0x2D};
    if (startsWith(info, signature, 5)) return "PDF";

    signature = {0x50, 0x4B, 0x03, 0x04};
    if (startsWith(info, signature, 4)) return "ZIP";

    signature = {0x1F, 0x8B};
    if (startsWith(info, signature, 2)) return "GZIP";

    signature = {0x52, 0x61, 0x72, 0x21, 0x1A, 0x07};
    if (startsWith(info, signature, 6)) return "RAR";

    signature = {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C};
    if (startsWith(info, signature, 6)) return "7-Zip";

    signature = {0x42, 0x4D};
    if (startsWith(info, signature, 2)) return "BMP";

    return "Unknown";
}

} // namespace cmdr::file
