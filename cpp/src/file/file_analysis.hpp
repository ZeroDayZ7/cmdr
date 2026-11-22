#pragma once

#include "file/extension_check.hpp"
#include "file/file_format.hpp"
#include "file/file_inspector.hpp"

#include <filesystem>
#include <string>

namespace cmdr::file {

struct FileAnalysisResult {
    bool success{false};
    std::string errorMessage;
    FileInfo info;
    FileFormat format{FileFormat::Unknown};
    ExtensionCheck extensionCheck{};
};

FileAnalysisResult analyzeFile(const std::filesystem::path& path);

} // namespace cmdr::file
