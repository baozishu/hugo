import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/api_service.dart';
import 'package:typecho_blog_client/services/auth_service.dart';

class PostDetailScreen extends StatefulWidget {
  final int postId;

  const PostDetailScreen({Key? key, required this.postId}) : super(key: key);

  @override
  _PostDetailScreenState createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  Post? _post;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadPostDetail();
  }

  Future<void> _loadPostDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final apiService = ApiService(Provider.of<AuthService>(context, listen: false));
      _post = await apiService.getPostDetail(widget.postId);
    } catch (e) {
      setState(() {
        _errorMessage = '加载文章失败: $e';
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('文章详情'),
        actions: [
          IconButton(
            icon: Icon(Icons.share),
            onPressed: () {
              // 实现分享功能
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text('分享功能待实现')),
              );
            },
          ),
        ],
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
                        onPressed: _loadPostDetail,
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : _post != null
                  ? SingleChildScrollView(
                      padding: EdgeInsets.all(16.0),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _post!.title,
                            style: Theme.of(context).textTheme.headline4,
                          ),
                          SizedBox(height: 16),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                '作者: ${_post!.author}',
                                style: TextStyle(color: Colors.grey),
                              ),
                              Text(
                                _post!.date,
                                style: TextStyle(color: Colors.grey),
                              ),
                            ],
                          ),
                          SizedBox(height: 8),
                          Row(
                            children: [
                              Chip(
                                label: Text(_post!.category),
                                backgroundColor: Theme.of(context).accentColor.withOpacity(0.2),
                              ),
                              SizedBox(width: 8),
                              if (_post!.tags.isNotEmpty)
                                Expanded(
                                  child: Wrap(
                                    spacing: 8,
                                    children: _post!.tags
                                        .map(
                                          (tag) => Chip(
                                            label: Text(tag),
                                            backgroundColor: Theme.of(context).primaryColor.withOpacity(0.2),
                                            labelStyle: TextStyle(fontSize: 12),
                                          ),
                                        )
                                        .toList(),
                                  ),
                                ),
                            ],
                          ),
                          SizedBox(height: 24),
                          Markdown(
                            data: _post!.content,
                            physics: NeverScrollableScrollPhysics(),
                            padding: EdgeInsets.zero,
                            selectable: true,
                          ),
                        ],
                      ),
                    )
                  : Center(child: Text('文章不存在')),
    );
  }
}
