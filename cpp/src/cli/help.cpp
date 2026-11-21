#include "cli/help.hpp"

#include <iostream>

namespace cmdr::cli {

void printHelp() {
    std::cout
        << "Usage: cmdr-inspector [options]\n"
        << "\n"
        << "Options:\n"
        << "  -h, --help    Show this help message\n";
}

} // namespace cmdr::cli
