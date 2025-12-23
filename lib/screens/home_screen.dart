import 'package:flutter/material.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/components/post_list_item.dart';
import 'package:typecho_blog_client/providers/auth_provider.dart';
import 'package:provider/provider.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<Post> _posts = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  int _currentPage = 1;
  String? _errorMessage;
  bool _isRefreshing = false;
  late TypechoApiService _apiService;

  @override
  void initState() {
    super.initState();
    _apiService = context.read<AuthProvider>().apiService;
    _loadInitialPosts();
  }

  // 加载初始文章列表
  void _loadInitialPosts() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    
    try {
      final posts = await _apiService.getPosts(page: 1);
      setState(() {
        _posts = posts;
        _currentPage = 1;
        _hasMore = posts.length == 10; // 假设每页10篇文章
      });
    } catch (e) {
      _handleError('加载文章列表失败，请下拉刷新重试');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  // 下拉刷新
  Future<void> _refreshPosts() async {
    if (_isRefreshing) return;
    
    setState(() {
      _isRefreshing = true;
      _errorMessage = null;
    });
    
    try {
      final posts = await _apiService.getPosts(page: 1);
      setState(() {
        _posts = posts;
        _currentPage = 1;
        _hasMore = posts.length == 10;
      });
    } catch (e) {
      _handleError('刷新失败，当前显示缓存内容');
    } finally {
      setState(() {
        _isRefreshing = false;
      });
    }
  }

  // 加载更多文章
  Future<void> _loadMorePosts() async {
    if (_isLoadingMore || !_hasMore) return;
    
    setState(() {
      _isLoadingMore = true;
    });
    
    try {
      final morePosts = await _apiService.getPosts(page: _currentPage + 1);
      if (morePosts.isEmpty) {
        setState(() {
          _hasMore = false;
        });
      } else {
        setState(() {
          _posts.addAll(morePosts);
          _currentPage++;
          _hasMore = morePosts.length == 10;
        });
      }
    } catch (e) {
      _handleError('加载更多失败，请点击重试');
    } finally {
      setState(() {
        _isLoadingMore = false;
      });
    }
  }

  // 处理错误情况
  void _handleError(String message) {
    setState(() {
      _errorMessage = message;
    });
    
    // 显示错误提示
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        action: SnackBarAction(
          label: '重试',
          onPressed: () => _refreshPosts(),
        ),
        duration: Duration(seconds: 3),
      ),
    );
  }

  // 滚动监听
  void _scrollListener() {
    ScrollPosition position = Scrollable.of(context)!.position;
    if (position.pixels == position.maxScrollExtent && !_isLoadingMore && _hasMore) {
      _loadMorePosts();
    }
  }

  // 构建加载指示器
  Widget _buildLoadingIndicator() {
    return Container(
      padding: EdgeInsets.all(16.0),
      alignment: Alignment.center,
      child: _isLoadingMore
          ? Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                CircularProgressIndicator(),
                SizedBox(height: 8.0),
                Text('加载更多中...'),
              ],
            )
          : !_hasMore
              ? Text('已经到底啦')
              : Container(),
    );
  }

  // 构建错误提示界面
  Widget _buildErrorView() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.error_outline, size: 64, color: Colors.red),
          SizedBox(height: 16.0),
          Text(
            _errorMessage ?? '加载失败',
            style: TextStyle(fontSize: 16.0, color: Colors.grey[700]),
            textAlign: TextAlign.center,
          ),
          SizedBox(height: 24.0),
          ElevatedButton(
            onPressed: _refreshPosts,
            child: Text('重新加载'),
          ),
        ],
      ),
    );
  }

  // 构建主界面
  Widget _buildBody() {
    if (_isLoading) {
      return Center(child: CircularProgressIndicator());
    }
    
    if (_errorMessage != null) {
      return _buildErrorView();
    }
    
    return RefreshIndicator(
      onRefresh: _refreshPosts,
      child: NotificationListener<ScrollNotification>(
        onNotification: (ScrollNotification scrollInfo) {
          if (scrollInfo.metrics.pixels == scrollInfo.metrics.maxScrollExtent &&
              !_isLoadingMore &&
              _hasMore) {
            _loadMorePosts();
          }
          return false;
        },
        child: ListView.builder(
          itemCount: _posts.length + 1, // +1 for loading indicator
          itemBuilder: (context, index) {
            if (index == _posts.length) {
              return _buildLoadingIndicator();
            }
            
            final post = _posts[index];
            return PostListItem(
              post: post,
              onTap: () {
                // 导航到文章详情页
                Navigator.pushNamed(
                  context,
                  '/postDetail',
                  arguments: post.id,
                );
              },
            );
          },
          physics: AlwaysScrollableScrollPhysics(), // 即使内容不足一屏也可以下拉刷新
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('博客'),
        actions: [
          IconButton(
            icon: Icon(Icons.search),
            onPressed: () {
              Navigator.pushNamed(context, '/search');
            },
          ),
        ],
      ),
      body: _buildBody(),
    );
  }
}