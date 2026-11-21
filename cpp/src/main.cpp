#include "cli/help.hpp"

#include <iostream>
#include <string_view>

int main(int argc, char* argv[]) {
    if (argc == 1) {
        cmdr::cli::printHelp();
        return 0;
    }

    if (argc != 2) {
        std::cerr << "Too many arguments.\n";
        std::cerr << "Run 'cmdr-inspector --help' for usage information.\n";
        return 2;
    }

    const std::string_view argument(argv[1]);

    if (argument == "-h" || argument == "--help") {
        cmdr::cli::printHelp();
        return 0;
    }

    std::cerr << "Unknown argument: " << argument << '\n';
    std::cerr << "Run 'cmdr-inspector --help' for usage information.\n";
    return 2;
}
