#include "file/extension_check.hpp"

#include <algorithm>
#include <array>
#include <cctype>
#include <initializer_list>
#include <string>

namespace cmdr::file {

namespace {

std::string toLowerAscii(std::string value) {
    std::transform(value.begin(), value.end(), value.begin(), [](unsigned char character) {
        return static_cast<char>(std::tolower(character));
    });
    return value;
}

bool containsExtension(
    const std::string& extension,
    std::initializer_list<const char*> knownExtensions
) {
    for (const char* known : knownExtensions) {
        if (extension == known) {
            return true;
        }
    }
    return false;
}

bool isKnownExtension(const std::string& extension) {
    return containsExtension(extension, {
        ".elf", ".exe", ".com", ".dll", ".sys", ".drv",
        ".png", ".jpg", ".jpeg", ".jpe",
        ".gif", ".gif87a", ".gif89a",
        ".pdf", ".zip", ".gz", ".gzip", ".tgz",
        ".rar", ".7z", ".bmp"
    });
}

} // namespace

ExtensionCheck checkExtension(
    const std::filesystem::path& path,
    std::string_view detectedFormat
) {
    std::filesystem::path extensionPath = path.extension();
    if (extensionPath.empty()) {
        return ExtensionCheck{std::string{}, ExtensionStatus::NoExtension};
    }

    std::string extension = toLowerAscii(extensionPath.string());

    if (detectedFormat == "Unknown") {
        return ExtensionCheck{extension, ExtensionStatus::UnknownFormat};
    }

    if (detectedFormat == "ELF") {
        if (containsExtension(extension, {".elf"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "DOS/Windows MZ executable") {
        if (containsExtension(extension, {".exe", ".com", ".dll", ".sys", ".drv"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "PNG") {
        if (containsExtension(extension, {".png"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "JPEG") {
        if (containsExtension(extension, {".jpg", ".jpeg", ".jpe"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "GIF87a" || detectedFormat == "GIF89a") {
        if (containsExtension(extension, {".gif", ".gif87a", ".gif89a"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "PDF") {
        if (containsExtension(extension, {".pdf"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "ZIP") {
        if (containsExtension(extension, {".zip"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "GZIP") {
        if (containsExtension(extension, {".gz", ".gzip", ".tgz"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "RAR") {
        if (containsExtension(extension, {".rar"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "7-Zip") {
        if (containsExtension(extension, {".7z"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    } else if (detectedFormat == "BMP") {
        if (containsExtension(extension, {".bmp"})) {
            return ExtensionCheck{extension, ExtensionStatus::Match};
        }
    }

    if (isKnownExtension(extension)) {
        return ExtensionCheck{extension, ExtensionStatus::KnownMismatch};
    }

    return ExtensionCheck{extension, ExtensionStatus::UnknownExtension};
}

} // namespace cmdr::file
