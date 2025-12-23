import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/providers/theme_provider.dart';
import 'package:typecho_blog_client/screens/home_screen.dart';
import 'package:typecho_blog_client/screens/login_screen.dart';
import 'package:typecho_blog_client/screens/post_detail_screen.dart';
import 'package:typecho_blog_client/screens/settings_screen.dart';
import 'package:typecho_blog_client/screens/edit_post_screen.dart';
import 'package:typecho_blog_client/screens/my_posts_screen.dart';
import 'package:typecho_blog_client/screens/categories_screen.dart';
import 'package:typecho_blog_client/screens/tags_screen.dart';
import 'package:typecho_blog_client/screens/search_screen.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (context) => ThemeProvider()),
        Provider(create: (context) => AuthService()),
      ],
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final themeProvider = Provider.of<ThemeProvider>(context);
    final authService = Provider.of<AuthService>(context);
    
    return MaterialApp(
      title: 'Typecho博客客户端',
      theme: themeProvider.lightTheme,
      darkTheme: themeProvider.darkTheme,
      themeMode: themeProvider.themeMode,
      // 添加全局过渡动画
      pageTransitionsTheme: PageTransitionsTheme(
        builders: {
          TargetPlatform.android: CustomPageTransitionsBuilder(),
          TargetPlatform.iOS: CustomPageTransitionsBuilder(),
          TargetPlatform.fuchsia: CustomPageTransitionsBuilder(),
          TargetPlatform.linux: CustomPageTransitionsBuilder(),
          TargetPlatform.macOS: CustomPageTransitionsBuilder(),
          TargetPlatform.windows: CustomPageTransitionsBuilder(),
        },
      ),
      home: FutureBuilder<bool>(
        future: authService.loadSavedCredentials(),
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return Scaffold(
              body: Center(child: CircularProgressIndicator()),
            );
          }
          return snapshot.data == true ? HomeScreen() : LoginScreen();
        },
      ),
      onGenerateRoute: (settings) {
        return _generateRoute(context, settings);
      },
    );
  }
  
  Route<dynamic>? _generateRoute(BuildContext context, RouteSettings settings) {
    final authService = Provider.of<AuthService>(context, listen: false);
    
    // 路由守卫
    if (!authService.isAuthenticated && settings.name != '/login') {
      return CustomPageRoute(builder: (context) => LoginScreen());
    }
    
    switch (settings.name) {
      case '/':
        return CustomPageRoute(builder: (context) => HomeScreen());
      case '/login':
        return CustomPageRoute(builder: (context) => LoginScreen());
      case '/home':
        return CustomPageRoute(builder: (context) => HomeScreen());
      case '/post_detail':
        final postId = settings.arguments as int?;
        return CustomPageRoute(
          builder: (context) => PostDetailScreen(postId: postId),
        );
      case '/settings':
        return CustomPageRoute(builder: (context) => SettingsScreen());
      case '/edit_post':
        final postId = settings.arguments as int?;
        return CustomPageRoute(
          builder: (context) => EditPostScreen(postId: postId),
        );
      case '/my_posts':
        return CustomPageRoute(builder: (context) => MyPostsScreen());
      case '/categories':
        return CustomPageRoute(builder: (context) => CategoriesScreen());
      case '/tags':
        return CustomPageRoute(builder: (context) => TagsScreen());
      case '/search':
        final query = settings.arguments as String?;
        return CustomPageRoute(
          builder: (context) => SearchScreen(initialQuery: query),
        );
      default:
        return CustomPageRoute(builder: (context) => HomeScreen());
    }
  }
}

// 自定义页面路由实现平滑过渡动画
class CustomPageRoute<T> extends PageRoute<T> {
  final WidgetBuilder builder;
  final Curve curve;
  final Duration duration;
  
  CustomPageRoute({
    required this.builder,
    this.curve = Curves.easeInOut,
    this.duration = const Duration(milliseconds: 300),
  });

  @override
  Color? get barrierColor => null;

  @override
  String? get barrierLabel => null;

  @override
  bool get maintainState => true;

  @override
  Duration get transitionDuration => duration;

  @override
  Widget buildPage(
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
  ) {
    final curvedAnimation = CurvedAnimation(
      parent: animation,
      curve: curve,
    );
    
    return FadeTransition(
      opacity: curvedAnimation,
      child: SlideTransition(
        position: Tween<Offset>(
          begin: const Offset(0.1, 0),
          end: Offset.zero,
        ).animate(curvedAnimation),
        child: builder(context),
      ),
    );
  }
}

// 自定义页面过渡构建器
class CustomPageTransitionsBuilder extends PageTransitionsBuilder {
  @override
  Widget buildTransitions<T>(
    PageRoute<T> route,
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
    Widget child,
  ) {
    const Duration duration = Duration(milliseconds: 300);
    
    // 进入动画
    if (route.isFirst) {
      return child;
    }
    
    // 自定义过渡动画
    return FadeTransition(
      opacity: animation,
      child: SlideTransition(
        position: Tween<Offset>(
          begin: const Offset(0.1, 0),
          end: Offset.zero,
        ).animate(
          CurvedAnimation(
            parent: animation,
            curve: Curves.easeInOut,
            reverseCurve: Curves.easeInOut,
          ),
        ),
        child: child,
      ),
    );
  }
}