#include <array>
#include <cstddef>
#include <cstdlib>
#include <filesystem>
#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>

#ifdef _WIN32
#include <windows.h>
#else
#include <cerrno>
#include <cstring>
#include <sys/wait.h>
#include <unistd.h>
#endif

namespace {

struct ProcessResult {
    int exitCode;
    std::string stdoutText;
    std::string stderrText;
};

void require(bool condition, const char* message) {
    if (!condition) {
        throw std::runtime_error(message);
    }
}

bool contains(const std::string& text, const std::string& fragment) {
    return text.find(fragment) != std::string::npos;
}

std::string quoteWindowsArgument(const std::string& argument) {
    if (argument.find_first_of(" \\\"\t") == std::string::npos) {
        return argument;
    }

    std::string result = "\"";
    std::size_t backslashes = 0;

    for (const char character : argument) {
        if (character == '\\') {
            ++backslashes;
        } else {
            if (character == '"') {
                result.append(backslashes * 2 + 1, '\\');
            }
            result.append(backslashes, '\\');
            backslashes = 0;
            result.push_back(character);
        }
    }

    result.append(backslashes * 2, '\\');
    result.push_back('"');
    return result;
}

#ifdef _WIN32

std::wstring toWide(const std::string& text) {
    if (text.empty()) {
        return std::wstring();
    }

    const int size = MultiByteToWideChar(
        CP_UTF8,
        MB_ERR_INVALID_CHARS,
        text.c_str(),
        -1,
        nullptr,
        0
    );
    if (size <= 0) {
        throw std::runtime_error("Could not convert argument to UTF-16");
    }

    std::wstring result(static_cast<std::size_t>(size), L'\0');
    if (MultiByteToWideChar(
            CP_UTF8,
            MB_ERR_INVALID_CHARS,
            text.c_str(),
            -1,
            result.data(),
            size
        ) <= 0) {
        throw std::runtime_error("Could not convert argument to UTF-16");
    }
    result.pop_back();
    return result;
}

ProcessResult runProcess(
    const std::filesystem::path& executable,
    const std::vector<std::string>& arguments
) {
    SECURITY_ATTRIBUTES security{};
    security.nLength = sizeof(security);
    security.bInheritHandle = TRUE;

    HANDLE stdoutRead = nullptr;
    HANDLE stdoutWrite = nullptr;
    HANDLE stderrRead = nullptr;
    HANDLE stderrWrite = nullptr;

    if (!CreatePipe(&stdoutRead, &stdoutWrite, &security, 0) ||
        !SetHandleInformation(stdoutRead, HANDLE_FLAG_INHERIT, 0)) {
        throw std::runtime_error("Could not create stdout pipe");
    }
    if (!CreatePipe(&stderrRead, &stderrWrite, &security, 0) ||
        !SetHandleInformation(stderrRead, HANDLE_FLAG_INHERIT, 0)) {
        CloseHandle(stdoutRead);
        CloseHandle(stdoutWrite);
        throw std::runtime_error("Could not create stderr pipe");
    }

    std::string commandLine;
    commandLine += quoteWindowsArgument(executable.string());
    for (const std::string& argument : arguments) {
        commandLine.push_back(' ');
        commandLine += quoteWindowsArgument(argument);
    }

    std::wstring application = toWide(executable.string());
    std::wstring command = toWide(commandLine);

    STARTUPINFOW startup{};
    startup.cb = sizeof(startup);
    startup.dwFlags = STARTF_USESTDHANDLES;
    startup.hStdOutput = stdoutWrite;
    startup.hStdError = stderrWrite;

    PROCESS_INFORMATION process{};

    if (!CreateProcessW(
            application.c_str(),
            command.data(),
            nullptr,
            nullptr,
            TRUE,
            0,
            nullptr,
            nullptr,
            &startup,
            &process
        )) {
        CloseHandle(stdoutRead);
        CloseHandle(stdoutWrite);
        CloseHandle(stderrRead);
        CloseHandle(stderrWrite);
        throw std::runtime_error("Could not start cmdr-inspector");
    }

    CloseHandle(stdoutWrite);
    CloseHandle(stderrWrite);

    std::string stdoutText;
    std::string stderrText;
    std::array<char, 4096> buffer{};

    DWORD bytesRead = 0;
    while (ReadFile(stdoutRead, buffer.data(),
                    static_cast<DWORD>(buffer.size()), &bytesRead, nullptr) &&
           bytesRead > 0) {
        stdoutText.append(buffer.data(), bytesRead);
    }

    while (ReadFile(stderrRead, buffer.data(),
                    static_cast<DWORD>(buffer.size()), &bytesRead, nullptr) &&
           bytesRead > 0) {
        stderrText.append(buffer.data(), bytesRead);
    }

    WaitForSingleObject(process.hProcess, INFINITE);

    DWORD exitCode = 0;
    GetExitCodeProcess(process.hProcess, &exitCode);

    CloseHandle(process.hThread);
    CloseHandle(process.hProcess);
    CloseHandle(stdoutRead);
    CloseHandle(stderrRead);

    return ProcessResult{
        static_cast<int>(exitCode), stdoutText, stderrText
    };
}

#else

std::string readPipe(int descriptor) {
    std::string result;
    std::array<char, 4096> buffer{};

    for (;;) {
        const ssize_t count = read(descriptor, buffer.data(), buffer.size());
        if (count > 0) {
            result.append(buffer.data(), static_cast<std::size_t>(count));
        } else if (count == 0) {
            break;
        } else if (errno != EINTR) {
            close(descriptor);
            throw std::runtime_error("Could not read child process pipe");
        }
    }

    close(descriptor);
    return result;
}

ProcessResult runProcess(
    const std::filesystem::path& executable,
    const std::vector<std::string>& arguments
) {
    int stdoutPipe[2];
    int stderrPipe[2];

    if (pipe(stdoutPipe) != 0 || pipe(stderrPipe) != 0) {
        throw std::runtime_error("Could not create process pipes");
    }

    const pid_t child = fork();
    if (child < 0) {
        throw std::runtime_error("Could not fork process");
    }

    if (child == 0) {
        close(stdoutPipe[0]);
        close(stderrPipe[0]);
        dup2(stdoutPipe[1], STDOUT_FILENO);
        dup2(stderrPipe[1], STDERR_FILENO);
        close(stdoutPipe[1]);
        close(stderrPipe[1]);

        std::vector<char*> argv;
        argv.push_back(const_cast<char*>(executable.c_str()));
        for (const std::string& argument : arguments) {
            argv.push_back(const_cast<char*>(argument.c_str()));
        }
        argv.push_back(nullptr);

        execv(executable.c_str(), argv.data());
        _exit(127);
    }

    close(stdoutPipe[1]);
    close(stderrPipe[1]);

    std::string stdoutText = readPipe(stdoutPipe[0]);
    std::string stderrText = readPipe(stderrPipe[0]);

    int status = 0;
    waitpid(child, &status, 0);

    int exitCode = 0;
    if (WIFEXITED(status)) {
        exitCode = WEXITSTATUS(status);
    } else {
        exitCode = 127;
    }

    return ProcessResult{exitCode, stdoutText, stderrText};
}

#endif

void requireExit(const ProcessResult& result, int expected) {
    require(result.exitCode == expected,
            ("Unexpected exit code: " + std::to_string(result.exitCode)).c_str());
}

void requireStdout(const ProcessResult& result, const std::string& fragment) {
    require(contains(result.stdoutText, fragment),
            ("stdout missing: " + fragment).c_str());
}

void requireStderr(const ProcessResult& result, const std::string& fragment) {
    require(contains(result.stderrText, fragment),
            ("stderr missing: " + fragment).c_str());
}

} // namespace

int main(int argc, char* argv[]) {
    try {
        require(argc == 3,
                "cli_tests expects executable path and samples directory");

        const std::filesystem::path executable(argv[1]);
        const std::filesystem::path samples(argv[2]);

        ProcessResult result = runProcess(executable, {"--help"});
        requireExit(result, 0);
        requireStdout(result, "Usage: cmdr-inspector");

        result = runProcess(executable, {"-h"});
        requireExit(result, 0);
        requireStdout(result, "Options:");

        result = runProcess(executable, {});
        requireExit(result, 0);
        requireStdout(result, "Usage: cmdr-inspector");

        result = runProcess(executable, {"file", (samples / "sample.png").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: PNG");

        result = runProcess(executable, {"file", (samples / "sample.pdf").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: PDF");

        result = runProcess(executable, {"file", (samples / "sample.bin").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: Unknown");

        result = runProcess(executable, {"file", (samples / "empty.dat").string()});
        requireExit(result, 0);
        requireStdout(result, "Size: 0 bytes");
        requireStdout(result, "Detected format: Unknown");

        result = runProcess(executable,
                            {"file", (samples / "with spaces" / "sample with spaces.txt").string()});
        requireExit(result, 0);
        requireStdout(result, "File name: sample with spaces.txt");

        result = runProcess(executable,
                            {"file", (samples / "missing.dat").string()});
        requireExit(result, 1);
        requireStderr(result, "Path does not exist");

        result = runProcess(executable, {"file", samples.string()});
        requireExit(result, 1);
        requireStderr(result, "Path is not a regular file");

        result = runProcess(executable, {"unknown"});
        requireExit(result, 2);
        requireStderr(result, "Unknown argument: unknown");

        result = runProcess(executable, {"file", "one", "two"});
        requireExit(result, 2);
        requireStderr(result, "Too many arguments");

        result = runProcess(executable, {"file", (samples / "png-as-jpg.jpg").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: PNG");
        requireStdout(result, "Extension: .jpg");
        requireStdout(result, "Warning: Extension does not match detected format.");

        result = runProcess(executable, {"file", (samples / "sample.webp").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: WebP");
        requireStdout(result, "Extension: .webp");

        result = runProcess(executable, {"file", (samples / "sample.wav").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: WAV");
        requireStdout(result, "Extension: .wav");

        result = runProcess(executable, {"file", (samples / "sample.ogg").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: OGG");

        result = runProcess(executable, {"file", (samples / "sample.flac").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: FLAC");

        result = runProcess(executable, {"file", (samples / "sample.sqlite").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: SQLite");
        requireStdout(result, "Extension: .sqlite");

        result = runProcess(executable, {"file", (samples / "riff-only.bin").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: Unknown");

        result = runProcess(executable, {"file", (samples / "riff-wrong.bin").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: Unknown");

        result = runProcess(executable, {"file", (samples / "png-unknown.xyz").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: PNG");
        requireStdout(result, "Extension: .xyz");

        result = runProcess(executable, {"file", (samples / "noextension").string()});
        requireExit(result, 0);
        requireStdout(result, "Detected format: PNG");
        requireStdout(result, "Extension: (none)");

        std::cout << "cli_tests: all integration checks passed\n";
        return EXIT_SUCCESS;
    } catch (const std::exception& error) {
        std::cerr << "cli_tests: " << error.what() << '\n';
        return EXIT_FAILURE;
    }
}
