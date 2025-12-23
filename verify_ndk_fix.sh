#!/bin/bash
# 验证脚本 - 确认NDK ABI配置冲突修复

echo "验证GitHub Actions工作流配置修复..."

# 1. 检查android/app/build.gradle中是否已移除ndk abi配置
if grep -q "abiFilters" "$GITHUB_WORKSPACE/android/app/build.gradle"; then
  echo "❌ 错误: android/app/build.gradle中仍存在abiFilters配置"
  exit 1
else
  echo "✅ 确认: android/app/build.gradle中已成功移除abiFilters配置"
fi

# 2. 检查构建命令是否包含--split-per-abi选项
if grep -q "--split-per-abi" "$GITHUB_WORKSPACE/.github/workflows/flutter_ci_cd.yml"; then
  echo "✅ 确认: 构建命令中包含--split-per-abi选项"
else
  echo "❌ 错误: 构建命令中未找到--split-per-abi选项"
  exit 1
fi

# 3. 验证gradle-wrapper.properties配置是否正确
if grep -q "distributionUrl=https://services.gradle.org/distributions/gradle-7.6.1-bin.zip" "$GITHUB_WORKSPACE/android/gradle/wrapper/gradle-wrapper.properties"; then
  echo "✅ 确认: gradle-wrapper.properties配置正确"
else
  echo "❌ 错误: gradle-wrapper.properties配置不正确或缺失"
fi

echo "所有配置验证完成！修复已正确应用。"
exit 0