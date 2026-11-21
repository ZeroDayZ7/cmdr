#pragma once

#include <filesystem>
#include <string>
#include <string_view>

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
    std::string_view detectedFormat
);

} // namespace cmdr::file
