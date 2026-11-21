#include "file/extension_check.hpp"

#include <cstdlib>
#include <filesystem>
#include <iostream>
#include <stdexcept>
#include <string>

namespace {

void require(bool condition, const char* message) {
    if (!condition) {
        throw std::runtime_error(message);
    }
}

void requireStatus(
    const std::filesystem::path& path,
    std::string_view format,
    cmdr::file::ExtensionStatus expected
) {
    const cmdr::file::ExtensionCheck result =
        cmdr::file::checkExtension(path, format);
    require(result.status == expected, "Unexpected extension status");
}

} // namespace

int main() {
    try {
        requireStatus("photo.png", "PNG", cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.PNG", "PNG", cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.jpg", "PNG", cmdr::file::ExtensionStatus::KnownMismatch);

        requireStatus("photo.jpg", "JPEG", cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.jpeg", "JPEG", cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.pdf", "PDF", cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.webp", "WebP", cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.wav", "WAV", cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.ogg", "OGG", cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.flac", "FLAC", cmdr::file::ExtensionStatus::Match);
        requireStatus("database.sqlite", "SQLite", cmdr::file::ExtensionStatus::Match);
        requireStatus("database.db", "SQLite", cmdr::file::ExtensionStatus::Match);

        requireStatus("photo.png", "WAV", cmdr::file::ExtensionStatus::KnownMismatch);

        requireStatus("photo", "PNG", cmdr::file::ExtensionStatus::NoExtension);
        requireStatus("photo.custom", "PNG", cmdr::file::ExtensionStatus::UnknownExtension);
        requireStatus("photo.unknown", "Unknown", cmdr::file::ExtensionStatus::UnknownFormat);

        const cmdr::file::ExtensionCheck jpgPng =
            cmdr::file::checkExtension("photo.JPG", "PNG");
        require(jpgPng.extension == ".jpg", "PNG mismatch extension was not lowercase");

        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "extension_check_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
