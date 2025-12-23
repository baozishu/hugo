import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/screens/edit_post_screen.dart';
import 'package:typecho_blog_client/components/post_list_item.dart';
import 'package:typecho_blog_client/components/custom_button.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class MyPostsScreen extends StatefulWidget {
  @override
  _MyPostsScreenState createState() => _MyPostsScreenState();
}

class _MyPostsScreenState extends State> MyPostsScreen> {
  List>Post> _posts = [];
  bool _isLoading = true;
  bool _isRefreshing = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadPosts();
  }

  Future>void> _loadPosts() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of>AuthService>(context, listen: false));
      _posts = await apiService.getMyPosts();
    } catch (e) {
      setState(() {
        _errorMessage = '加载文章失败: $e';
      });
    } finally {
      setState(() {
        _isLoading = false;
        _isRefreshing = false;
      });
    }
  }

  Future>void> _refresh() async {
    setState(() {
      _isRefreshing = true;
    });
    await _loadPosts();
  }

  void _confirmDelete(String postId) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('确认删除'),
        content: Text('确定要删除这篇文章吗？此操作不可撤销。'),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
            },
            child: Text('取消'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _deletePost(postId);
            },
            child: Text('删除', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }

  Future>void> _deletePost(String postId) async {
    try {
      final apiService = TypechoApiService(Provider.of>AuthService>(context, listen: false));
      await apiService.deletePost(postId);
      
      // 重新加载文章列表
      _loadPosts();
      
      // 显示成功消息
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('文章删除成功')),
      );
    } catch (e) {
      // 显示错误消息
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('删除失败: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('我的文章'),
        actions: [
          CustomButton(
            onPressed: () async {
              // 导航到创建文章页面
              final result = await Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => EditPostScreen()),
              );
              
              // 如果创建成功，刷新文章列表
              if (result == true) {
                _loadPosts();
              }
            },
            text: '创建',
            icon: Icon(Icons.add),
          ),
        ],
      ),
      body: _isLoading && !_isRefreshing
          ? Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_errorMessage!),
                      ElevatedButton(
                        onPressed: _loadPosts,
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _refresh,
                  child: _posts.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.article, size: 64, color: Theme.of(context).disabledColor),
                              SizedBox(height: AppTheme.mediumPadding),
                              Text('您还没有发布过文章'),
                              SizedBox(height: AppTheme.mediumPadding),
                              ElevatedButton(
                                onPressed: () async {
                                  final result = await Navigator.push(
                                    context,
                                    MaterialPageRoute(builder: (context) => EditPostScreen()),
                                  );
                                  
                                  if (result == true) {
                                    _loadPosts();
                                  }
                                },
                                child: Text('创建第一篇文章'),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: _posts.length,
                          itemBuilder: (context, index) {
                            final post = _posts[index];
                            return Card(
                              margin: EdgeInsets.symmetric(
                                horizontal: AppTheme.smallPadding,
                                vertical: AppTheme.smallPadding / 2,
                              ),
                              child: ListTile(
                                title: Text(post.title),
                                subtitle: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '发布时间: ${post.date ?? ''}',
                                      style: TextStyle(fontSize: 12),
                                    ),
                                    SizedBox(height: 4),
                                    Text(
                                      post.excerpt?.substring(0, post.excerpt.length > 100 ? 100 : post.excerpt.length) ?? '',
                                      style: TextStyle(fontSize: 12, color: Theme.of(context).textTheme.caption?.color),
                                      maxLines: 2,
                                    ),
                                  ],
                                ),
                                trailing: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      icon: Icon(Icons.edit),
                                      onPressed: () async {
                                        // 导航到编辑页面
                                        final result = await Navigator.push(
                                          context,
                                          MaterialPageRoute(builder: (context) => EditPostScreen(postId: post.id)),
                                        );
                                        
                                        // 如果编辑成功，刷新列表
                                        if (result == true) {
                                          _loadPosts();
                                        }
                                      },
                                    ),
                                    IconButton(
                                      icon: Icon(Icons.delete, color: Colors.red),
                                      onPressed: () {
                                        _confirmDelete(post.id);
                                      },
                                    ),
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
                ),
    );
  }
}
