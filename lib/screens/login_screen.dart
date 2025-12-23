import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/services/auth_service.dart';

class LoginScreen extends StatefulWidget {
  @override
  _LoginScreenState createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final TextEditingController _apiUrlController = TextEditingController(text: 'https://your-blog-url.com');
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _tryAutoLogin();
  }

  Future<void> _tryAutoLogin() async {
    final authService = Provider.of<AuthService>(context, listen: false);
    bool isLoggedIn = await authService.loadSavedCredentials();
    
    if (isLoggedIn) {
      Navigator.pushReplacementNamed(context, '/home');
    }
  }

  Future<void> _handleLogin() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final authService = Provider.of<AuthService>(context, listen: false);
    bool success = await authService.login(
      _apiUrlController.text,
      _usernameController.text,
      _passwordController.text,
    );

    setState(() {
      _isLoading = false;
    });

    if (success) {
      Navigator.pushReplacementNamed(context, '/home');
    } else {
      setState(() {
        _errorMessage = '登录失败，请检查您的API地址、用户名和密码';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: EdgeInsets.all(20.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                FlutterLogo(size: 100),
                SizedBox(height: 30),
                Text(
                  'Typecho博客客户端',
                  style: Theme.of(context).textTheme.headline6,
                ),
                SizedBox(height: 30),
                TextField(
                  controller: _apiUrlController,
                  decoration: InputDecoration(
                    labelText: '博客API地址',
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.link),
                  ),
                  keyboardType: TextInputType.url,
                ),
                SizedBox(height: 16),
                TextField(
                  controller: _usernameController,
                  decoration: InputDecoration(
                    labelText: '用户名',
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.person),
                  ),
                ),
                SizedBox(height: 16),
                TextField(
                  controller: _passwordController,
                  decoration: InputDecoration(
                    labelText: '密码',
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.lock),
                  ),
                  obscureText: true,
                ),
                if (_errorMessage != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 16.0),
                    child: Text(
                      _errorMessage!, 
                      style: TextStyle(color: Colors.red),
                    ),
                  ),
                SizedBox(height: 30),
                ElevatedButton(
                  onPressed: _isLoading ? null : _handleLogin,
                  child: _isLoading ? CircularProgressIndicator() : Text('登录'),
                  style: ElevatedButton.styleFrom(
                    padding: EdgeInsets.symmetric(horizontal: 40, vertical: 12),
                  ),
                ),
                SizedBox(height: 20),
                TextButton(
                  onPressed: () {
                    // 显示帮助信息
                    showDialog(
                      context: context, 
                      builder: (context) => AlertDialog(
                        title: Text('使用说明'),
                        content: Text('请输入您的Typecho博客地址、用户名和密码。\n\n注意：确保您的博客已开启API功能，API地址通常为您的博客地址。'),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(context), 
                            child: Text('确定'),
                          ),
                        ],
                      ),
                    );
                  },
                  child: Text('使用帮助'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
