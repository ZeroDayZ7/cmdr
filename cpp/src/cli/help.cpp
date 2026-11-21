#include "cli/help.hpp"

#include <iostream>

namespace cmdr::cli {

void printHelp() {
    std::cout
        << "Usage: cmdr-inspector [options]\n"
        << "       cmdr-inspector file <path>\n"
        << "\n"
        << "Options:\n"
        << "  -h, --help    Show this help message\n"
        << "\n"
        << "Commands:\n"
        << "  file <path>   Inspect basic file information\n";
}

} // namespace cmdr::cli
