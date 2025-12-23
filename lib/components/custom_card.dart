import 'package:flutter/material.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class CustomCard extends StatelessWidget {
  final Widget? child;
  final EdgeInsets? padding;
  final Function()? onTap;
  final bool withShadow;
  final double? borderRadius;

  const CustomCard({
    Key? key,
    this.child,
    this.padding,
    this.onTap,
    this.withShadow = true,
    this.borderRadius,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Card(
      elevation: withShadow ? 4 : 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(borderRadius ?? AppTheme.mediumRadius),
      ),
      color: isDarkMode ? AppTheme.darkCard : AppTheme.lightCard,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(borderRadius ?? AppTheme.mediumRadius),
        child: Padding(
          padding: padding ?? AppTheme.mediumPadding,
          child: child,
        ),
      ),
    );
  }
}
