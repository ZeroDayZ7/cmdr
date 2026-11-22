#pragma once

#include "file/extension_check.hpp"
#include "file/file_inspector.hpp"

#include <string_view>

namespace cmdr::file {

FileFormat detectFileType(const FileInfo& info);
std::string_view fileFormatName(FileFormat format);

} // namespace cmdr::file
