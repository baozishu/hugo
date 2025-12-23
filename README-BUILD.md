# Hugo Visual Client 构建说明

## 构建要求

Hugo Visual Client 使用 Fyne GUI 框架，需要以下依赖：

### Windows 构建要求

1. **Go 1.21+** - 已安装
2. **CGO 支持** - 需要 C 编译器
3. **TDM-GCC 或 MinGW-w64** - 用于 CGO 编译

#### 安装 TDM-GCC (推荐)

1. 下载 TDM-GCC: https://jmeubank.github.io/tdm-gcc/
2. 安装到默认路径 (C:\TDM-GCC-64)
3. 确保 PATH 环境变量包含 C:\TDM-GCC-64\bin

#### 或者安装 MinGW-w64

```bash
# 使用 Chocolatey
choco install mingw

# 或者使用 Scoop
scoop install mingw
```

### 验证安装

```bash
gcc --version
go env CGO_ENABLED
```

## 构建命令

### 快速构建 (当前平台)

```bash
go build -o hugo-visual-client.exe cmd/hugo-client/main.go
```

### 优化构建 (减小文件大小)

```bash
go build -ldflags "-s -w -H windowsgui" -o hugo-visual-client.exe cmd/hugo-client/main.go
```

### 跨平台构建

使用提供的批处理文件：

```bash
build-client.bat
```

## 运行应用程序

构建完成后，直接运行可执行文件：

```bash
./hugo-visual-client.exe
```

## 故障排除

### CGO 错误

如果遇到 CGO 相关错误：

1. 确保安装了 C 编译器 (TDM-GCC 或 MinGW)
2. 检查 PATH 环境变量
3. 重启命令行窗口
4. 验证 `go env CGO_ENABLED` 返回 "1"

### OpenGL 错误

如果遇到 OpenGL 相关错误：

1. 更新显卡驱动
2. 确保系统支持 OpenGL 3.2+

### 依赖问题

如果遇到依赖问题：

```bash
go mod tidy
go mod download
```

## 打包分发

构建的可执行文件是独立的，可以直接分发。但需要注意：

1. 目标系统需要支持 OpenGL
2. Windows 版本需要 Windows 7+ 
3. 文件大小约 20-30MB (包含 GUI 库)

## 开发模式运行

开发时可以直接运行：

```bash
go run cmd/hugo-client/main.go
```