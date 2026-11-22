#include "file/file_analysis.hpp"

#include "file/file_type.hpp"

namespace cmdr::file {

FileAnalysisResult analyzeFile(const std::filesystem::path& path) {
    FileAnalysisResult result;

    if (!inspectFile(path, result.info, result.errorMessage)) {
        return result;
    }

    result.format = detectFileType(result.info);
    result.extensionCheck = checkExtension(path, result.format);
    result.success = true;
    result.errorMessage.clear();
    return result;
}

} // namespace cmdr::file
