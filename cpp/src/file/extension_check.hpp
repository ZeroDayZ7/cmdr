#pragma once

#include <filesystem>
#include <string>

namespace cmdr::file {

enum class FileFormat {
    Unknown,
    ELF,
    DOSWindowsMZ,
    PNG,
    JPEG,
    GIF87a,
    GIF89a,
    PDF,
    ZIP,
    GZIP,
    RAR,
    SevenZip,
    BMP,
    WebP,
    WAV,
    OGG,
    FLAC,
    SQLite
};

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
