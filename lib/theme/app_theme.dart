import 'package:flutter/material.dart';

class AppTheme {
  // 亮色主题颜色
  static const Color lightPrimary = Color(0xFF1976D2);
  static const Color lightSecondary = Color(0xFF03A9F4);
  static const Color lightBackground = Color(0xFFFFFFFF);
  static const Color lightCard = Color(0xFFFFFFFF);
  static const Color lightText = Color(0xFF212121);
  static const Color lightTextSecondary = Color(0xFF757575);
  static const Color lightDivider = Color(0xFFEEEEEE);
  
  // 暗色主题颜色
  static const Color darkPrimary = Color(0xFFBBDEFB);
  static const Color darkSecondary = Color(0xFF80D8FF);
  static const Color darkBackground = Color(0xFF121212);
  static const Color darkCard = Color(0xFF1E1E1E);
  static const Color darkText = Color(0xFFFFFFFF);
  static const Color darkTextSecondary = Color(0xFFB0B0B0);
  static const Color darkDivider = Color(0xFF2C2C2C);

  // 通用颜色
  static const Color error = Color(0xFFB00020);
  static const Color success = Color(0xFF4CAF50);
  static const Color warning = Color(0xFFFB8C00);
  static const Color info = Color(0xFF2196F3);

  // 间距
  static const EdgeInsets smallPadding = EdgeInsets.all(8.0);
  static const EdgeInsets mediumPadding = EdgeInsets.all(16.0);
  static const EdgeInsets largePadding = EdgeInsets.all(24.0);

  // 圆角
  static const double smallRadius = 4.0;
  static const double mediumRadius = 8.0;
  static const double largeRadius = 16.0;

  // 字体大小
  static const double smallFontSize = 12.0;
  static const double normalFontSize = 14.0;
  static const double mediumFontSize = 16.0;
  static const double largeFontSize = 18.0;
  static const double xlargeFontSize = 20.0;
  static const double xxlargeFontSize = 24.0;

  // 阴影
  static const BoxShadow smallShadow = BoxShadow(
    color: Color(0x22000000),
    spreadRadius: 2,
    blurRadius: 4,
    offset: Offset(0, 2),
  );

  static const BoxShadow mediumShadow = BoxShadow(
    color: Color(0x22000000),
    spreadRadius: 4,
    blurRadius: 8,
    offset: Offset(0, 4),
  );
}
