#pragma once

#include "file/file_format.hpp"

#include <filesystem>
#include <string>

namespace cmdr::file {

enum class ExtensionStatus {
    Match,
    KnownMismatch,
    UnknownExtension,
    NoExtension,
    UnknownFormat
};

struct ExtensionCheck {
    std::string extension;
    ExtensionStatus status;
};

ExtensionCheck checkExtension(
    const std::filesystem::path& path,
    FileFormat detectedFormat
);

} // namespace cmdr::file
