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
    cmdr::file::FileFormat format,
    cmdr::file::ExtensionStatus expected
) {
    const cmdr::file::ExtensionCheck result =
        cmdr::file::checkExtension(path, format);
    require(result.status == expected, "Unexpected extension status");
}

} // namespace

int main() {
    try {
        requireStatus("program.elf", cmdr::file::FileFormat::ELF, cmdr::file::ExtensionStatus::Match);
        requireStatus("program.exe", cmdr::file::FileFormat::DOSWindowsMZ, cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.png", cmdr::file::FileFormat::PNG, cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.PNG", cmdr::file::FileFormat::PNG, cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.jpg", cmdr::file::FileFormat::JPEG, cmdr::file::ExtensionStatus::Match);
        requireStatus("photo.jpeg", cmdr::file::FileFormat::JPEG, cmdr::file::ExtensionStatus::Match);
        requireStatus("animation.gif87a", cmdr::file::FileFormat::GIF87a, cmdr::file::ExtensionStatus::Match);
        requireStatus("animation.gif89a", cmdr::file::FileFormat::GIF89a, cmdr::file::ExtensionStatus::Match);
        requireStatus("document.pdf", cmdr::file::FileFormat::PDF, cmdr::file::ExtensionStatus::Match);
        requireStatus("archive.zip", cmdr::file::FileFormat::ZIP, cmdr::file::ExtensionStatus::Match);
        requireStatus("archive.gz", cmdr::file::FileFormat::GZIP, cmdr::file::ExtensionStatus::Match);
        requireStatus("archive.rar", cmdr::file::FileFormat::RAR, cmdr::file::ExtensionStatus::Match);
        requireStatus("archive.7z", cmdr::file::FileFormat::SevenZip, cmdr::file::ExtensionStatus::Match);
        requireStatus("image.bmp", cmdr::file::FileFormat::BMP, cmdr::file::ExtensionStatus::Match);
        requireStatus("image.webp", cmdr::file::FileFormat::WebP, cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.wav", cmdr::file::FileFormat::WAV, cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.ogg", cmdr::file::FileFormat::OGG, cmdr::file::ExtensionStatus::Match);
        requireStatus("audio.flac", cmdr::file::FileFormat::FLAC, cmdr::file::ExtensionStatus::Match);
        requireStatus("database.sqlite", cmdr::file::FileFormat::SQLite, cmdr::file::ExtensionStatus::Match);
        requireStatus("database.db", cmdr::file::FileFormat::SQLite, cmdr::file::ExtensionStatus::Match);

        requireStatus("photo.jpg", cmdr::file::FileFormat::PNG, cmdr::file::ExtensionStatus::KnownMismatch);
        requireStatus("photo.png", cmdr::file::FileFormat::WAV, cmdr::file::ExtensionStatus::KnownMismatch);

        requireStatus("photo", cmdr::file::FileFormat::PNG, cmdr::file::ExtensionStatus::NoExtension);
        requireStatus("photo.custom", cmdr::file::FileFormat::PNG, cmdr::file::ExtensionStatus::UnknownExtension);
        requireStatus("photo.unknown", cmdr::file::FileFormat::Unknown, cmdr::file::ExtensionStatus::UnknownFormat);

        const cmdr::file::ExtensionCheck jpgPng =
            cmdr::file::checkExtension("photo.JPG", cmdr::file::FileFormat::PNG);
        require(jpgPng.extension == ".jpg", "PNG mismatch extension was not lowercase");

        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "extension_check_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
