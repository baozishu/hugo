import 'package:flutter/material.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class CustomButton extends StatelessWidget {
  final String text;
  final Function()? onPressed;
  final Color? backgroundColor;
  final Color? textColor;
  final double? borderRadius;
  final double? paddingVertical;
  final double? paddingHorizontal;
  final bool isLoading;
  final bool isOutline;

  const CustomButton({
    Key? key,
    required this.text,
    this.onPressed,
    this.backgroundColor,
    this.textColor,
    this.borderRadius,
    this.paddingVertical = 12,
    this.paddingHorizontal = 24,
    this.isLoading = false,
    this.isOutline = false,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return ElevatedButton(
      onPressed: isLoading ? null : onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: isOutline
            ? Colors.transparent
            : (backgroundColor ?? (isDarkMode ? AppTheme.lightPrimary : AppTheme.lightPrimary)),
        foregroundColor: textColor ?? Colors.white,
        padding: EdgeInsets.symmetric(
          vertical: paddingVertical!, 
          horizontal: paddingHorizontal!,
        ),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(borderRadius ?? AppTheme.mediumRadius),
          side: isOutline
              ? BorderSide(
                  color: backgroundColor ?? (isDarkMode ? AppTheme.lightPrimary : AppTheme.lightPrimary),
                )
              : BorderSide.none,
        ),
      ),
      child: isLoading
          ? SizedBox(
              width: 20,
              height: 20,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                color: isOutline ? (backgroundColor ?? Theme.of(context).primaryColor) : Colors.white,
              ),
            )
          : Text(
              text,
              style: TextStyle(
                fontWeight: FontWeight.w600,
                fontSize: AppTheme.mediumFontSize,
                color: isOutline
                    ? (backgroundColor ?? (isDarkMode ? AppTheme.lightPrimary : AppTheme.lightPrimary))
                    : (textColor ?? Colors.white),
              ),
            ),
    );
  }
}
