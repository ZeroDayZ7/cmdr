#include "cli/help.hpp"
#include "file/file_analysis.hpp"
#include "file/file_type.hpp"

#include <filesystem>
#include <iomanip>
#include <iostream>
#include <string>
#include <string_view>

int main(int argc, char* argv[]) {
    if (argc == 1) {
        cmdr::cli::printHelp();
        return 0;
    }

    if (argc == 2) {
        const std::string_view argument(argv[1]);

        if (argument == "-h" || argument == "--help") {
            cmdr::cli::printHelp();
            return 0;
        }

        std::cerr << "Unknown argument: " << argument << '\n';
        std::cerr << "Run 'cmdr-inspector --help' for usage information.\n";
        return 2;
    }

    if (argc == 3) {
        const std::string_view command(argv[1]);

        if (command == "file") {
            const std::filesystem::path path(argv[2]);
            const cmdr::file::FileAnalysisResult result =
                cmdr::file::analyzeFile(path);

            if (!result.success) {
                std::cerr << result.errorMessage << '\n';
                return 1;
            }

            const cmdr::file::FileInfo& info = result.info;

            std::cout << "File name: " << info.name << '\n';
            std::cout << "Size: " << info.sizeBytes << " bytes\n";
            std::cout << "First 16 bytes (HEX):\n";

            for (std::size_t index = 0; index < info.bytesRead; ++index) {
                if (index > 0) {
                    std::cout << ' ';
                }

                std::cout << std::uppercase << std::hex
                          << std::setw(2) << std::setfill('0')
                          << static_cast<unsigned int>(info.firstBytes[index]);
            }

            std::cout << '\n';
            std::cout << "Detected format: "
                      << cmdr::file::fileFormatName(result.format) << '\n';

            if (result.extensionCheck.status == cmdr::file::ExtensionStatus::NoExtension) {
                std::cout << "Extension: (none)\n";
            } else {
                std::cout << "Extension: " << result.extensionCheck.extension << '\n';
            }

            if (result.extensionCheck.status == cmdr::file::ExtensionStatus::Match) {
                std::cout << "Extension matches detected format.\n";
            } else if (result.extensionCheck.status == cmdr::file::ExtensionStatus::KnownMismatch) {
                std::cout << "Warning: Extension does not match detected format.\n";
            } else if (result.extensionCheck.status == cmdr::file::ExtensionStatus::UnknownExtension) {
                std::cout << "Extension: unknown for detected format.\n";
            }

            return 0;
        }

        std::cerr << "Unknown argument: " << command << '\n';
        std::cerr << "Run 'cmdr-inspector --help' for usage information.\n";
        return 2;
    }

    std::cerr << "Too many arguments.\n";
    std::cerr << "Run 'cmdr-inspector --help' for usage information.\n";
    return 2;
}
