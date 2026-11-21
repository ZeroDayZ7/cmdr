#pragma once

#include "file/file_inspector.hpp"

#include <string_view>

namespace cmdr::file {

std::string_view detectFileType(const FileInfo& info);

} // namespace cmdr::file
