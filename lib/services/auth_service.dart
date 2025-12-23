import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

class AuthService {
  String? _apiUrl;
  String? _username;
  String? _password;
  bool _isAuthenticated = false;

  bool get isAuthenticated =\u003e _isAuthenticated;
  String? get apiUrl =\u003e _apiUrl;

  Future\u003cbool\u003e login(String apiUrl, String username, String password) async {
    _apiUrl = apiUrl;
    _username = username;
    _password = password;

    try {
      // 测试连接是否正常
      final response = await http.get(
        Uri.parse('$_apiUrl/action/api?json=1'),
        headers: {
          'Authorization': 'Basic \${base64Encode(utf8.encode('$_username:$_password'))}',
          'Accept': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        _isAuthenticated = true;
        await _saveCredentials();
        return true;
      } else {
        _isAuthenticated = false;
        return false;
      }
    } catch (e) {
      _isAuthenticated = false;
      return false;
    }
  }

  Future\u003cvoid\u003e logout() async {
    _isAuthenticated = false;
    _username = null;
    _password = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('api_url');
    await prefs.remove('username');
    await prefs.remove('password');
  }

  Future\u003cbool\u003e loadSavedCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    _apiUrl = prefs.getString('api_url');
    _username = prefs.getString('username');
    _password = prefs.getString('password');

    if (_apiUrl != null && _username != null && _password != null) {
      return await login(_apiUrl!, _username!, _password!);
    }

    return false;
  }

  Future\u003cvoid\u003e _saveCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('api_url', _apiUrl!);
    await prefs.setString('username', _username!);
    await prefs.setString('password', _password!);
  }

  Map\u003cString, String\u003e getAuthHeaders() {
    return {
      'Authorization': 'Basic \${base64Encode(utf8.encode('$_username:$_password'))}',
      'Accept': 'application/json',
    };
  }
}
