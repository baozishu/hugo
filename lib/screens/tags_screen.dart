import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/components/post_list_item.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/screens/post_detail_screen.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class TagsScreen extends StatefulWidget {
  @override
  _TagsScreenState createState() => _TagsScreenState();
}

class _TagsScreenState extends State<TagsScreen> {
  List<Map<String, dynamic>> _tags = [];
  Map<String, dynamic>? _selectedTag;
  List<Post> _posts = [];
  bool _isLoading = true;
  bool _isLoadingPosts = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadTags();
  }

  Future<void> _loadTags() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of<AuthService>(context, listen: false));
      _tags = await apiService.getTags();
      
      if (_tags.isNotEmpty) {
        _selectTag(_tags[0]);
      }
    } catch (e) {
      setState(() {
        _errorMessage = '加载标签失败: $e';
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _selectTag(Map<String, dynamic> tag) async {
    setState(() {
      _selectedTag = tag;
      _isLoadingPosts = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of<AuthService>(context, listen: false));
      _posts = await apiService.getPosts(tag: tag['name']);
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
        title: Text('标签'),
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
                        onPressed: _loadTags,
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : Row(
                  children: [
                    // 左侧标签列表
                    Container(
                      width: 100,
                      child: ListView.builder(
                        itemCount: _tags.length,
                        itemBuilder: (context, index) {
                          final tag = _tags[index];
                          final isSelected = _selectedTag?['mid'] == tag['mid'];
                          
                          return GestureDetector(
                            onTap: () => _selectTag(tag),
                            child: Container(
                              padding: EdgeInsets.symmetric(vertical: 16, horizontal: 8),
                              decoration: BoxDecoration(
                                color: isSelected
                                    ? Theme.of(context).accentColor.withOpacity(0.1)
                                    : Colors.transparent,
                                borderLeft: BorderSide(
                                  color: isSelected
                                      ? Theme.of(context).accentColor
                                      : Colors.transparent,
                                  width: 3,
                                ),
                              ),
                              child: Text(
                                tag['name'],
                                textAlign: TextAlign.center,
                                style: TextStyle(
                                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                                  color: isSelected
                                      ? Theme.of(context).accentColor
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
                                          if (_selectedTag != null) {
                                            _selectTag(_selectedTag!);
                                          }
                                        },
                                        child: Text('重试'),
                                      ),
                                    ],
                                  ),
                                )
                              : _posts.isEmpty
                                  ? Center(child: Text('该标签下暂无文章'))
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
