#include "file/file_inspector.hpp"

#include <cstdint>
#include <cstdlib>
#include <filesystem>
#include <fstream>
#include <iostream>
#include <stdexcept>
#include <string>

namespace {

void require(bool condition, const char* message) {
    if (!condition) {
        throw std::runtime_error(message);
    }
}

class TemporaryDirectory {
public:
    TemporaryDirectory()
        : path_(std::filesystem::temp_directory_path()
                / ("cmdr-inspector-test-" + std::to_string(std::rand()))) {
        std::filesystem::create_directories(path_);
    }

    ~TemporaryDirectory() {
        std::error_code error;
        std::filesystem::remove_all(path_, error);
    }

    const std::filesystem::path& path() const {
        return path_;
    }

private:
    std::filesystem::path path_;
};

void writeFile(
    const std::filesystem::path& path,
    const std::string& content
) {
    std::ofstream output(path, std::ios::binary);
    require(output.is_open(), "Could not create temporary test file");
    output.write(content.data(), static_cast<std::streamsize>(content.size()));
    require(!output.fail(), "Could not write temporary test file");
}

void writeBinaryFile(
    const std::filesystem::path& path,
    const std::uint8_t* data,
    std::size_t size
) {
    std::ofstream output(path, std::ios::binary);
    require(output.is_open(), "Could not create binary test file");
    output.write(reinterpret_cast<const char*>(data),
                 static_cast<std::streamsize>(size));
    require(!output.fail(), "Could not write binary test file");
}

} // namespace

int main() {
    try {
        TemporaryDirectory temporary;

        const std::filesystem::path existing = temporary.path() / "existing.txt";
        writeFile(existing, "abc");

        cmdr::file::FileInfo info;
        std::string error;
        require(cmdr::file::inspectFile(existing, info, error),
                "Existing file should be inspected");
        require(info.name == "existing.txt", "Unexpected file name");
        require(info.sizeBytes == 3, "Unexpected file size");
        require(info.bytesRead == 3, "Short file should read all bytes");

        const std::filesystem::path empty = temporary.path() / "empty.bin";
        writeFile(empty, "");
        info = cmdr::file::FileInfo{};
        require(cmdr::file::inspectFile(empty, info, error),
                "Empty file should be inspected");
        require(info.sizeBytes == 0, "Empty file size must be zero");
        require(info.bytesRead == 0, "Empty file must have no bytesRead");

        const std::uint8_t binary[] = {0x00, 0x7F, 0x80, 0xFF, 0x10};
        const std::filesystem::path binaryPath = temporary.path() / "binary.bin";
        writeBinaryFile(binaryPath, binary, sizeof(binary));
        info = cmdr::file::FileInfo{};
        require(cmdr::file::inspectFile(binaryPath, info, error),
                "Binary file should be inspected");
        require(info.bytesRead == sizeof(binary),
                "Binary file shorter than 16 bytes must read all bytes");
        require(info.firstBytes[0] == 0x00, "Binary byte 0 was not preserved");
        require(info.firstBytes[2] == 0x80, "Binary byte 2 was not preserved");
        require(info.firstBytes[3] == 0xFF, "Binary byte 3 was not preserved");

        const std::filesystem::path shortPath = temporary.path() / "short.bin";
        writeBinaryFile(shortPath, binary, 5);
        info = cmdr::file::FileInfo{};
        require(cmdr::file::inspectFile(shortPath, info, error),
                "Short file should be inspected");
        require(info.bytesRead == 5, "Short file must report 5 bytes");

        const std::filesystem::path longPath = temporary.path() / "long.bin";
        std::array<std::uint8_t, 32> longContent{};
        for (std::size_t index = 0; index < longContent.size(); ++index) {
            longContent[index] = static_cast<std::uint8_t>(index);
        }
        writeBinaryFile(longPath, longContent.data(), longContent.size());
        info = cmdr::file::FileInfo{};
        require(cmdr::file::inspectFile(longPath, info, error),
                "Long file should be inspected");
        require(info.sizeBytes == 32, "Long file size must be 32");
        require(info.bytesRead == 16, "Only 16 bytes should be read");

        const std::filesystem::path missing = temporary.path() / "missing.bin";
        info = cmdr::file::FileInfo{};
        require(!cmdr::file::inspectFile(missing, info, error),
                "Missing path must fail");
        require(!error.empty(), "Missing path must provide an error message");

        info = cmdr::file::FileInfo{};
        require(!cmdr::file::inspectFile(temporary.path(), info, error),
                "Directory must not be treated as a regular file");
        require(!error.empty(), "Directory must provide an error message");

        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "file_inspector_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
