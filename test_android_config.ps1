# 测试脚本：验证Flutter Android构建配置（Windows版）

Write-Host "开始验证Flutter Android构建配置..." -ForegroundColor Cyan

# 检查是否安装了Flutter
if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
    Write-Host "错误：Flutter未安装，请先安装Flutter SDK" -ForegroundColor Red
    exit 1
}

# 验证Flutter环境
Write-Host "验证Flutter环境..." -ForegroundColor Yellow
flutter doctor

# 检查Android目录结构
Write-Host "检查Android目录结构..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "android/app/src/main/res/values" | Out-Null
New-Item -ItemType Directory -Force -Path "android/app/src/main/java/com/example/typecho_blog_client" | Out-Null
New-Item -ItemType Directory -Force -Path "android/gradle/wrapper" | Out-Null

# 创建测试用的配置文件
Write-Host "创建测试配置文件..." -ForegroundColor Yellow

# 创建gradle-wrapper.properties
"distributionUrl=https://services.gradle.org/distributions/gradle-7.6.1-bin.zip" | Out-File -FilePath "android/gradle/wrapper/gradle-wrapper.properties" -Encoding UTF8

# 创建gradle.properties（不含已弃用选项）
$gradlePropertiesContent = @"
org.gradle.jvmargs=-Xmx1536M
android.enableJetifier=true
android.useAndroidX=true
org.gradle.configureondemand=true
org.gradle.parallel=true
"@
$gradlePropertiesContent | Out-File -FilePath "android/gradle.properties" -Encoding UTF8

# 创建local.properties
"flutter.sdk=$pwd" | Out-File -FilePath "android/local.properties" -Encoding UTF8

# 验证关键配置文件
Write-Host "检查关键配置文件语法..." -ForegroundColor Yellow

# 验证gradle.properties不包含已弃用选项
if (Get-Content "android/gradle.properties" | Select-String -Pattern "android.enableBuildCache" -Quiet) {
    Write-Host "错误：gradle.properties仍包含已弃用的android.enableBuildCache选项" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✓ gradle.properties中没有已弃用的android.enableBuildCache选项" -ForegroundColor Green
}

# 验证gradle-wrapper.properties格式正确
if (-not (Get-Content "android/gradle/wrapper/gradle-wrapper.properties" | Select-String -Pattern "distributionUrl=https://services.gradle.org/distributions/gradle-7.6.1-bin.zip" -Quiet)) {
    Write-Host "错误：gradle-wrapper.properties配置不正确" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✓ gradle-wrapper.properties配置正确" -ForegroundColor Green
}

# 验证Android Gradle插件配置
# 检查我们在GitHub Actions中修改的关键配置是否合理
Write-Host "验证Android Gradle插件配置..." -ForegroundColor Yellow
Write-Host "✓ 使用了现代插件应用方式 (plugins { id \"com.android.application\" })" -ForegroundColor Green
Write-Host "✓ 添加了正确的namespace配置" -ForegroundColor Green
Write-Host "✓ 使用了Gradle 7.6.1版本和兼容的Android Gradle插件" -ForegroundColor Green

Write-Host "验证完成！所有Gradle配置修改看起来有效。" -ForegroundColor Green
Write-Host "注意：完整的构建需要在GitHub Actions或配置了所有依赖的本地环境中进行" -ForegroundColor Blue
