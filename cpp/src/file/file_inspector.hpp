#pragma once

#include <array>
#include <cstddef>
#include <cstdint>
#include <filesystem>
#include <string>

namespace cmdr::file {

struct FileInfo {
    std::string name;
    std::uintmax_t sizeBytes;
    std::array<std::uint8_t, 16> firstBytes{};
    std::size_t bytesRead{0};
};

bool inspectFile(
    const std::filesystem::path& path,
    FileInfo& info,
    std::string& errorMessage
);

} // namespace cmdr::file
