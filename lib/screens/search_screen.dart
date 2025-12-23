import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/components/post_list_item.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/screens/post_detail_screen.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class SearchScreen extends StatefulWidget {
  @override
  _SearchScreenState createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  List<Post> _searchResults = [];
  TextEditingController _searchController = TextEditingController();
  bool _isSearching = false;
  bool _hasSearched = false;
  String? _errorMessage;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _search(String keyword) async {
    if (keyword.isEmpty) return;

    setState(() {
      _isSearching = true;
      _errorMessage = null;
      _hasSearched = true;
    });

    try {
      final apiService = TypechoApiService(Provider.of<AuthService>(context, listen: false));
      _searchResults = await apiService.searchPosts(keyword);
    } catch (e) {
      setState(() {
        _errorMessage = '搜索失败: $e';
      });
    } finally {
      setState(() {
        _isSearching = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Container(
          height: 40,
          decoration: BoxDecoration(
            color: Theme.of(context).cardColor,
            borderRadius: BorderRadius.circular(20),
          ),
          child: TextField(
            controller: _searchController,
            onSubmitted: _search,
            decoration: InputDecoration(
              hintText: '搜索文章...',
              contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              border: InputBorder.none,
              suffixIcon: IconButton(
                icon: Icon(Icons.search),
                onPressed: () => _search(_searchController.text),
              ),
            ),
          ),
        ),
      ),
      body: _isSearching
          ? Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_errorMessage!),
                      ElevatedButton(
                        onPressed: () => _search(_searchController.text),
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : !_hasSearched
                  ? Center(
                      child: Text('请输入关键词进行搜索'),
                    )
                  : _searchResults.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.search_off, size: 64, color: Theme.of(context).disabledColor),
                              SizedBox(height: 16),
                              Text('没有找到相关文章'),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: _searchResults.length,
                          itemBuilder: (context, index) {
                            final post = _searchResults[index];
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
    );
  }
}
