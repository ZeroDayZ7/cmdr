#include "file/file_analysis.hpp"

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
                / ("cmdr-file-analysis-test-" + std::to_string(std::rand()))) {
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

void writeBinaryFile(
    const std::filesystem::path& path,
    const std::uint8_t* data,
    std::size_t size
) {
    std::ofstream output(path, std::ios::binary);
    require(output.is_open(), "Could not create test file");
    output.write(reinterpret_cast<const char*>(data),
                 static_cast<std::streamsize>(size));
    require(!output.fail(), "Could not write test file");
}

} // namespace

int main() {
    try {
        TemporaryDirectory temporary;

        const std::uint8_t pngSignature[] = {
            0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A
        };

        const std::filesystem::path matching = temporary.path() / "photo.PNG";
        writeBinaryFile(matching, pngSignature, sizeof(pngSignature));

        const cmdr::file::FileAnalysisResult matchResult =
            cmdr::file::analyzeFile(matching);
        require(matchResult.success, "Existing PNG file should be analyzed");
        require(matchResult.info.name == "photo.PNG", "Unexpected inspected file name");
        require(matchResult.format == cmdr::file::FileFormat::PNG,
                "PNG signature should be recognized");
        require(matchResult.extensionCheck.status == cmdr::file::ExtensionStatus::Match,
                "PNG extension should match");
        require(matchResult.extensionCheck.extension == ".png",
                "PNG extension should be returned lowercase");

        const std::filesystem::path mismatching = temporary.path() / "photo.jpg";
        writeBinaryFile(mismatching, pngSignature, sizeof(pngSignature));

        const cmdr::file::FileAnalysisResult mismatchResult =
            cmdr::file::analyzeFile(mismatching);
        require(mismatchResult.success, "Existing mismatching file should be analyzed");
        require(mismatchResult.format == cmdr::file::FileFormat::PNG,
                "PNG signature should be recognized for mismatch case");
        require(mismatchResult.extensionCheck.status == cmdr::file::ExtensionStatus::KnownMismatch,
                "JPEG extension should be a known mismatch for PNG data");

        const std::filesystem::path missing = temporary.path() / "missing.png";
        const cmdr::file::FileAnalysisResult missingResult =
            cmdr::file::analyzeFile(missing);
        require(!missingResult.success, "Missing file should fail analysis");
        require(!missingResult.errorMessage.empty(),
                "Missing file should return the existing inspection error");

        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "file_analysis_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
