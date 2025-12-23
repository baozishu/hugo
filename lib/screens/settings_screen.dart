import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/providers/theme_provider.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/screens/my_posts_screen.dart';

class SettingsScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final themeProvider = Provider.of<ThemeProvider>(context);
    final authService = Provider.of<AuthService>(context);

    return Scaffold(
      appBar: AppBar(
        title: Text('设置'),
      ),
      body: ListView(
        children: [
          ListTile(
            leading: Icon(Icons.article),
            title: Text('我的文章'),
            trailing: Icon(Icons.chevron_right),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => MyPostsScreen()),
              );
            },
          ),
          Divider(),
          ListTile(
            leading: Icon(Icons.color_lens),
            title: Text('深色模式'),
            subtitle: Text('当前：${themeProvider.isDarkMode ? '开启' : '关闭'}'),
            trailing: Switch(
              value: themeProvider.isDarkMode,
              onChanged: (value) {
                themeProvider.toggleTheme();
              },
            ),
          ),
          Divider(),
          ListTile(
            leading: Icon(Icons.info),
            title: Text('关于应用'),
            onTap: () {
              showAboutDialog(
                context: context,
                applicationName: 'Typecho博客客户端',
                applicationVersion: '1.0.0',
                applicationLegalese: '© 2023 Typecho博客客户端',
                children: [
                  Text('一个使用Flutter开发的跨平台Typecho博客客户端'),
                ],
              );
            },
          ),
          Divider(),
          ListTile(
            leading: Icon(Icons.logout),
            title: Text('退出登录'),
            textColor: Colors.red,
            onTap: () {
              showDialog(
                context: context,
                builder: (context) => AlertDialog(
                  title: Text('退出登录'),
                  content: Text('确定要退出登录吗？'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(context),
                      child: Text('取消'),
                    ),
                    TextButton(
                      onPressed: () async {
                        await authService.logout();
                        Navigator.pushNamedAndRemoveUntil(
                          context,
                          '/login',
                          (route) => false,
                        );
                      },
                      child: Text('确定'),
                      style: TextButton.styleFrom(
                        primary: Colors.red,
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
