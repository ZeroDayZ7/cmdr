#pragma once

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

} // namespace cmdr::file
