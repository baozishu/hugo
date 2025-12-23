import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ThemeProvider extends ChangeNotifier {
  ThemeMode themeMode = ThemeMode.system;

  ThemeData get lightTheme => ThemeData(
        brightness: Brightness.light,
        primarySwatch: Colors.blue,
        primaryColor: Colors.blue,
        accentColor: Colors.blueAccent,
        scaffoldBackgroundColor: Colors.white,
        cardColor: Colors.white,
        textTheme: TextTheme(
          bodyText1: TextStyle(color: Colors.black87),
          bodyText2: TextStyle(color: Colors.black54),
          headline6: TextStyle(color: Colors.black),
        ),
      );

  ThemeData get darkTheme => ThemeData(
        brightness: Brightness.dark,
        primarySwatch: Colors.blue,
        primaryColor: Colors.blue.shade700,
        accentColor: Colors.blueAccent.shade400,
        scaffoldBackgroundColor: Color(0xFF121212),
        cardColor: Color(0xFF1E1E1E),
        textTheme: TextTheme(
          bodyText1: TextStyle(color: Colors.white),
          bodyText2: TextStyle(color: Colors.white70),
          headline6: TextStyle(color: Colors.white),
        ),
      );

  bool get isDarkMode {
    if (themeMode == ThemeMode.system) {
      return WidgetsBinding.instance.window.platformBrightness == Brightness.dark;
    }
    return themeMode == ThemeMode.dark;
  }

  Future<void> toggleTheme() async {
    if (themeMode == ThemeMode.light) {
      themeMode = ThemeMode.dark;
    } else {
      themeMode = ThemeMode.light;
    }
    notifyListeners();
    _saveThemePreferences();
  }

  Future<void> loadThemePreferences() async {
    final prefs = await SharedPreferences.getInstance();
    final themeIndex = prefs.getInt('theme_mode') ?? 0;
    themeMode = ThemeMode.values[themeIndex];
    notifyListeners();
  }

  Future<void> _saveThemePreferences() async {
    final prefs = await SharedPreferences.getInstance();
    prefs.setInt('theme_mode', themeMode.index);
  }
}
