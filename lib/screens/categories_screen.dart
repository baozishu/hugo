import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/components/post_list_item.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/screens/post_detail_screen.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class CategoriesScreen extends StatefulWidget {
  @override
  _CategoriesScreenState createState() => _CategoriesScreenState();
}

class _CategoriesScreenState extends State<CategoriesScreen> {
  List<Map<String, dynamic>> _categories = [];
  Map<String, dynamic>? _selectedCategory;
  List<Post> _posts = [];
  bool _isLoading = true;
  bool _isLoadingPosts = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadCategories();
  }

  Future<void> _loadCategories() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of<AuthService>(context, listen: false));
      _categories = await apiService.getCategories();
      
      if (_categories.isNotEmpty) {
        _selectCategory(_categories[0]);
      }
    } catch (e) {
      setState(() {
        _errorMessage = '加载分类失败: $e';
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _selectCategory(Map<String, dynamic> category) async {
    setState(() {
      _selectedCategory = category;
      _isLoadingPosts = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of<AuthService>(context, listen: false));
      _posts = await apiService.getPosts(category: category['slug']);
    } catch (e) {
      setState(() {
        _errorMessage = '加载文章失败: $e';
      });
    } finally {
      setState(() {
        _isLoadingPosts = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('分类'),
      ),
      body: _isLoading
          ? Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_errorMessage!),
                      ElevatedButton(
                        onPressed: _loadCategories,
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : Row(
                  children: [
                    // 左侧分类列表
                    Container(
                      width: 100,
                      child: ListView.builder(
                        itemCount: _categories.length,
                        itemBuilder: (context, index) {
                          final category = _categories[index];
                          final isSelected = _selectedCategory?['mid'] == category['mid'];
                          
                          return GestureDetector(
                            onTap: () => _selectCategory(category),
                            child: Container(
                              padding: EdgeInsets.symmetric(vertical: 16, horizontal: 8),
                              decoration: BoxDecoration(
                                color: isSelected
                                    ? Theme.of(context).primaryColor.withOpacity(0.1)
                                    : Colors.transparent,
                                borderLeft: BorderSide(
                                  color: isSelected
                                      ? Theme.of(context).primaryColor
                                      : Colors.transparent,
                                  width: 3,
                                ),
                              ),
                              child: Text(
                                category['name'],
                                textAlign: TextAlign.center,
                                style: TextStyle(
                                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                                  color: isSelected
                                      ? Theme.of(context).primaryColor
                                      : Theme.of(context).textTheme.bodyText1?.color,
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                    // 右侧文章列表
                    Expanded(
                      child: _isLoadingPosts
                          ? Center(child: CircularProgressIndicator())
                          : _errorMessage != null
                              ? Center(
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Text(_errorMessage!),
                                      ElevatedButton(
                                        onPressed: () {
                                          if (_selectedCategory != null) {
                                            _selectCategory(_selectedCategory!);
                                          }
                                        },
                                        child: Text('重试'),
                                      ),
                                    ],
                                  ),
                                )
                              : _posts.isEmpty
                                  ? Center(child: Text('该分类下暂无文章'))
                                  : ListView.builder(
                                      itemCount: _posts.length,
                                      itemBuilder: (context, index) {
                                        final post = _posts[index];
                                        return PostListItem(
                                          post: post,
                                          onTap: () {
                                            Navigator.push(
                                              context,
                                              MaterialPageRoute(
                                                builder: (context) => PostDetailScreen(postId: post.id),
                                              ),
                                            );
                                          },
                                        );
                                      },
                                    ),
                    ),
                  ],
                ),
    );
  }
}
