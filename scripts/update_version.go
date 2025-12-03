package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reVersionSemver      = regexp.MustCompile(`^\d+\.\d+\.\d+(?:\.\d+)?$`)
	reAppVersion         = regexp.MustCompile(`const AppVersion = "v[^"]+"`)
	reYAMLVersion        = regexp.MustCompile(`(\s+version:\s+)"[^"]+"(\s+# The application version)`)
	rePlistCFBundle      = regexp.MustCompile(`(<key>CFBundleVersion</key>\s*<string>)[^<]+(</string>)`)
	rePlistShortVersion  = regexp.MustCompile(`(<key>CFBundleShortVersionString</key>\s*<string>)[^<]+(</string>)`)
	reManifestAppVersion = regexp.MustCompile(`(<assemblyIdentity[^>]*name="com\.codeswitch\.app"[^>]*version=")[^"]+(")`)
	reNSISProductVersion = regexp.MustCompile(`(!define INFO_PRODUCTVERSION\s+")[^"]+(")`)
	reNFPMVersion        = regexp.MustCompile(`(^version:\s+")[^"]+(")`)
)

func main() {
	// 获取版本号：优先从环境变量，其次从 Git tag
	version := os.Getenv("VERSION")
	if version == "" {
		// 尝试从 Git tag 获取
		cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		output, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting Git tag: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please set VERSION environment variable or ensure you're in a Git repository with tags\n")
			os.Exit(1)
		}
		version = strings.TrimSpace(string(output))
		// 去掉 'v' 前缀（如果有）
		version = strings.TrimPrefix(version, "v")
	}

	if version == "" {
		fmt.Fprintf(os.Stderr, "Version cannot be empty\n")
		os.Exit(1)
	}

	if !reVersionSemver.MatchString(version) {
		fmt.Fprintf(os.Stderr, "Invalid version: %s (expect format like 1.2.3 or 1.2.3.4)\n", version)
		os.Exit(1)
	}

	fmt.Printf("Updating version to: %s\n", version)

	// 更新所有文件
	files := []struct {
		path    string
		updater func(string, string) error
	}{
		{"version_service.go", updateGoFile},
		{"build/config.yml", updateYAMLFile},
		{"build/darwin/Info.plist", updatePlistFile},
		{"build/darwin/Info.dev.plist", updatePlistFile},
		{"build/windows/info.json", updateWindowsInfoJSON},
		{"build/windows/wails.exe.manifest", updateManifestFile},
		{"build/windows/nsis/wails_tools.nsh", updateNSISFile},
		{"build/linux/nfpm/nfpm.yaml", updateNFPMYAML},
		{"cmd/updater/updater.exe.manifest", updateUpdaterManifest},
	}

	for _, file := range files {
		if err := file.updater(file.path, version); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", file.path, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Updated %s\n", file.path)
	}

	fmt.Printf("\n✅ All version files updated successfully!\n")
}

// 读-改-写辅助：只有内容变更才写回，避免无意义的 mtime 更新
func updateFile(path string, transform func([]byte) ([]byte, error)) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated, err := transform(original)
	if err != nil {
		return err
	}

	if bytes.Equal(original, updated) {
		return nil
	}

	return os.WriteFile(path, updated, 0644)
}

// 更新 Go 文件中的版本常量
func updateGoFile(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		newContent := reAppVersion.ReplaceAllString(string(content), fmt.Sprintf(`const AppVersion = "v%s"`, version))
		return []byte(newContent), nil
	})
}

// 更新 YAML 文件（build/config.yml）
func updateYAMLFile(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		// 匹配 version: "1.2.5" 格式（在 info 部分）
		newContent := reYAMLVersion.ReplaceAllString(string(content), fmt.Sprintf(`$1"%s"$2`, version))
		return []byte(newContent), nil
	})
}

// 更新 macOS Info.plist 文件
func updatePlistFile(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		contentStr := string(content)

		// 更新 CFBundleVersion
		contentStr = rePlistCFBundle.ReplaceAllString(contentStr, fmt.Sprintf(`${1}%s${2}`, version))

		// 更新 CFBundleShortVersionString
		contentStr = rePlistShortVersion.ReplaceAllString(contentStr, fmt.Sprintf(`${1}%s${2}`, version))

		return []byte(contentStr), nil
	})
}

// 更新 updater.exe.manifest 文件
func updateUpdaterManifest(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		// 将版本号转换为 4 部分格式 (major.minor.patch.0)
		versionParts := strings.Split(version, ".")
		for len(versionParts) < 4 {
			versionParts = append(versionParts, "0")
		}
		manifestVersion := strings.Join(versionParts[:4], ".")

		// 只匹配 assemblyIdentity 标签内的 version 属性
		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")
		inAssemblyIdentity := false
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.Contains(trimmedLine, "<assemblyIdentity") {
				inAssemblyIdentity = true
			}
			isXMLDeclaration := strings.HasPrefix(trimmedLine, "<?xml")
			if inAssemblyIdentity && !isXMLDeclaration && strings.HasPrefix(line, "    ") && strings.Contains(line, `version="`) {
				lines[i] = fmt.Sprintf(`    version="%s"`, manifestVersion)
			}
			if inAssemblyIdentity && strings.Contains(trimmedLine, "/>") {
				break
			}
		}
		newContent := strings.Join(lines, "\n")
		return []byte(newContent), nil
	})
}

// 更新 Windows info.json
func updateWindowsInfoJSON(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, err
		}

		// 更新 fixed.file_version
		if fixed, ok := data["fixed"].(map[string]interface{}); ok {
			fixed["file_version"] = version
		}

		// 更新 info.0000.ProductVersion
		if info, ok := data["info"].(map[string]interface{}); ok {
			if info0000, ok := info["0000"].(map[string]interface{}); ok {
				info0000["ProductVersion"] = version
			}
		}

		newContent, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return nil, err
		}

		return newContent, nil
	})
}

// 更新 Windows manifest 文件
func updateManifestFile(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		// 只更新 assemblyIdentity 中的 version，不更新 XML 声明和 Microsoft.Windows.Common-Controls
		newContent := reManifestAppVersion.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))
		return []byte(newContent), nil
	})
}

// 更新 NSIS 工具脚本
func updateNSISFile(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		// 匹配 !define INFO_PRODUCTVERSION "1.2.5" 格式（完整行）
		newContent := reNSISProductVersion.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))
		return []byte(newContent), nil
	})
}

// 更新 Linux nfpm.yaml
func updateNFPMYAML(path, version string) error {
	return updateFile(path, func(content []byte) ([]byte, error) {
		// 匹配 version: "1.2.5" 格式（完整行，避免匹配其他 version）
		newContent := reNFPMVersion.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))
		return []byte(newContent), nil
	})
}
