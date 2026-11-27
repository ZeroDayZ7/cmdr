#include "file/file_inspector.hpp"

#include <fstream>
#include <system_error>

namespace cmdr::file {

bool inspectFile(
    const std::filesystem::path& path,
    FileInfo& info,
    std::string& errorMessage
) {
    std::error_code error;

    if (!std::filesystem::exists(path, error)) {
        if (error) {
            errorMessage = "Cannot check path: " + error.message();
        } else {
            errorMessage = "Path does not exist: " + path.string();
        }
        return false;
    }

    const bool regular = std::filesystem::is_regular_file(path, error);
    if (error) {
        errorMessage = "Cannot check file type: " + error.message();
        return false;
    }

    if (!regular) {
        errorMessage = "Path is not a regular file: " + path.string();
        return false;
    }

    const std::uintmax_t size = std::filesystem::file_size(path, error);
    if (error) {
        errorMessage = "Cannot read file size: " + error.message();
        return false;
    }

    info.name = path.filename().string();
    info.sizeBytes = size;
    info.firstBytes.fill(0);
    info.bytesRead = 0;

    std::ifstream input(path, std::ios::binary);
    if (!input.is_open()) {
        errorMessage = "Cannot open file: " + path.string();
        return false;
    }

    input.read(reinterpret_cast<char*>(info.firstBytes.data()),
               static_cast<std::streamsize>(info.firstBytes.size()));

    info.bytesRead = static_cast<std::size_t>(input.gcount());

    if (input.bad()) {
        errorMessage = "Cannot read file contents: " + path.string();
        return false;
    }

    errorMessage.clear();
    return true;
}

} // namespace cmdr::file
