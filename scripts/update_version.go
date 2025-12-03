package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

// 更新 Go 文件中的版本常量
func updateGoFile(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 匹配 const AppVersion = "v1.2.5" 格式
	re := regexp.MustCompile(`const AppVersion = "v[^"]+"`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`const AppVersion = "v%s"`, version))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}

// 更新 YAML 文件（build/config.yml）
func updateYAMLFile(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 匹配 version: "1.2.5" 格式（在 info 部分）
	re := regexp.MustCompile(`(\s+version:\s+)"[^"]+"(\s+# The application version)`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`$1"%s"$2`, version))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}

// 更新 macOS Info.plist 文件
func updatePlistFile(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 更新 CFBundleVersion - 匹配 <key>CFBundleVersion</key> 后面的 <string>版本号</string>
	// 使用多行模式，确保匹配完整的 key-value 对
	re1 := regexp.MustCompile(`(<key>CFBundleVersion</key>\s*<string>)[^<]+(</string>)`)
	if re1.MatchString(contentStr) {
		contentStr = re1.ReplaceAllString(contentStr, fmt.Sprintf(`$1%s$2`, version))
	}

	// 更新 CFBundleShortVersionString - 匹配 <key>CFBundleShortVersionString</key> 后面的 <string>版本号</string>
	re2 := regexp.MustCompile(`(<key>CFBundleShortVersionString</key>\s*<string>)[^<]+(</string>)`)
	if re2.MatchString(contentStr) {
		contentStr = re2.ReplaceAllString(contentStr, fmt.Sprintf(`$1%s$2`, version))
	}

	return ioutil.WriteFile(path, []byte(contentStr), 0644)
}

// 更新 updater.exe.manifest 文件
func updateUpdaterManifest(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 将版本号转换为 4 部分格式 (major.minor.patch.0)
	// 例如 1.2.5 -> 1.2.5.0
	versionParts := strings.Split(version, ".")
	for len(versionParts) < 4 {
		versionParts = append(versionParts, "0")
	}
	manifestVersion := strings.Join(versionParts[:4], ".")

	// 匹配 <assemblyIdentity ... version="1.0.0.0" ...> 格式（支持多行）
	// 使用 (?s) 使 . 匹配换行符
	re := regexp.MustCompile(`(?s)(<assemblyIdentity[^>]*name="CodeSwitch\.Updater"[^>]*version=")[^"]+(")`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, manifestVersion))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}

// 更新 Windows info.json
func updateWindowsInfoJSON(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
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
		return err
	}

	return ioutil.WriteFile(path, newContent, 0644)
}

// 更新 Windows manifest 文件
func updateManifestFile(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 只更新 assemblyIdentity 中的 version，不更新 XML 声明和 Microsoft.Windows.Common-Controls
	// 匹配 <assemblyIdentity ... version="1.2.5" ...> 格式，但排除 Microsoft.Windows.Common-Controls
	re := regexp.MustCompile(`(<assemblyIdentity[^>]*name="com\.codeswitch\.app"[^>]*version=")[^"]+(")`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}

// 更新 NSIS 工具脚本
func updateNSISFile(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 匹配 !define INFO_PRODUCTVERSION "1.2.5" 格式（完整行）
	re := regexp.MustCompile(`(!define INFO_PRODUCTVERSION\s+")[^"]+(")`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}

// 更新 Linux nfpm.yaml
func updateNFPMYAML(path, version string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 匹配 version: "1.2.5" 格式（完整行，避免匹配其他 version）
	re := regexp.MustCompile(`(^version:\s+")[^"]+(")`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`$1%s$2`, version))

	return ioutil.WriteFile(path, []byte(newContent), 0644)
}
