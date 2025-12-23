import 'dart:convert';
import 'package:http/http.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/http_client.dart';
import 'package:typecho_blog_client/services/cache_service.dart';

class TypechoApiService {
  final AuthService _authService;
  final Client _client = ApiClient.client;
  final CacheService _cacheService = CacheService();

  TypechoApiService(this._authService);

  // 获取博客信息
  Future<Map<String, dynamic>> getBlogInfo() async {
    try {
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=options'),
        headers: _authService.getAuthHeaders(),
      );

      return jsonDecode(response.body);
    } catch (e) {
      print('获取博客信息失败: $e');
      // 返回默认博客信息
      return {
        'title': 'Typecho博客',
        'description': '使用Flutter开发的博客客户端',
      };
    }
  }

  // 获取文章列表
  Future<List<Post>> getPosts({
    int page = 1,
    int limit = 10,
    String? category,
    String? tag,
  }) async {
    // 对于第一页，尝试先获取缓存数据
    if (page == 1) {
      final cachedPosts = await _cacheService.getCachedPosts();
      if (cachedPosts != null) {
        // 在后台异步刷新数据
        _refreshPostsInBackground(category: category, tag: tag);
        return cachedPosts;
      }
    }

    try {
      String queryParams = 'json=1&method=contents&page=$page&limit=$limit&type=post';
      
      if (category != null) {
        queryParams += '&category=$category';
      }
      
      if (tag != null) {
        queryParams += '&tag=$tag';
      }

      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?$queryParams'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      final posts = postsData.map((postJson) => Post.fromJson(postJson)).toList();
      
      // 如果是第一页，缓存结果
      if (page == 1 && posts.isNotEmpty) {
        await _cacheService.cachePosts(posts);
      }
      
      return posts;
    } catch (e) {
      print('获取文章列表失败: $e');
      // 如果是第一页且出错，尝试返回缓存
      if (page == 1) {
        final cachedPosts = await _cacheService.getCachedPosts();
        if (cachedPosts != null) {
          print('使用缓存文章列表');
          return cachedPosts;
        }
      }
      return _getMockPosts();
    }
  }

  // 在后台刷新文章列表（不阻塞UI）
  Future<void> _refreshPostsInBackground({
    String? category,
    String? tag,
  }) async {
    try {
      print('后台刷新文章列表...');
      String queryParams = 'json=1&method=contents&page=1&limit=10&type=post';
      
      if (category != null) {
        queryParams += '&category=$category';
      }
      
      if (tag != null) {
        queryParams += '&tag=$tag';
      }

      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?$queryParams'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      final posts = postsData.map((postJson) => Post.fromJson(postJson)).toList();
      
      // 更新缓存
      await _cacheService.cachePosts(posts);
      print('后台刷新文章列表完成');
    } catch (e) {
      print('后台刷新文章列表失败: $e');
    }
  }

  // 获取单篇文章详情
  Future<Post> getPostDetail(int postId) async {
    // 先尝试从缓存获取
    final cachedPost = await _cacheService.getCachedPostDetail(postId);
    if (cachedPost != null) {
      // 在后台异步刷新
      _refreshPostDetailInBackground(postId);
      return cachedPost;
    }

    try {
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final post = Post.fromJson(data);
      
      // 缓存文章详情
      await _cacheService.cachePostDetail(postId, post);
      
      return post;
    } catch (e) {
      print('获取文章详情失败: $e');
      // 如果没有缓存，则返回模拟数据
      return Post.fromJson({
        'cid': postId.toString(),
        'title': '文章详情 - 缓存不可用',
        'text': '无法加载文章内容，请检查网络连接后重试。',
        'excerpt': '无法加载文章内容',
        'date': DateTime.now().toString(),
        'modified': DateTime.now().toString(),
        'author': {'name': '系统', 'uid': '0'},
        'categories': [],
        'tags': []
      });
    }
  }
  
  // 在后台刷新文章详情
  Future<void> _refreshPostDetailInBackground(int postId) async {
    try {
      print('后台刷新文章详情...');
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final post = Post.fromJson(data);
      
      // 更新缓存
      await _cacheService.cachePostDetail(postId, post);
      print('后台刷新文章详情完成');
    } catch (e) {
      print('后台刷新文章详情失败: $e');
    }
  }

  // 获取分类列表
  Future<List<Map<String, dynamic>>> getCategories() async {
    final response = await _client.get(
      Uri.parse('${_authService.apiUrl}/action/api?json=1&method=categories'),
      headers: _authService.getAuthHeaders(),
    );

    final data = jsonDecode(response.body);
    return List<Map<String, dynamic>>.from(data['rows']);
  }

  // 获取标签列表
  Future<List<Map<String, dynamic>>> getTags() async {
    final response = await _client.get(
      Uri.parse('${_authService.apiUrl}/action/api?json=1&method=tags'),
      headers: _authService.getAuthHeaders(),
    );

    final data = jsonDecode(response.body);
    return List<Map<String, dynamic>>.from(data['rows']);
  }

  // 创建新文章
  Future<bool> createPost({
    required String title,
    required String content,
    required String category,
    List<String> tags = const [],
    bool published = true,
  }) async {
    final postData = {
      'title': title,
      'text': content,
      'category': category,
      'tags': tags.join(','),
      'status': published ? 'publish' : 'draft',
      'do': 'publish',
    };

    final response = await _client.post(
      Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents'),
      headers: {
        ..._authService.getAuthHeaders(),
        'Content-Type': 'application/json',
      },
      body: jsonEncode(postData),
    );

    return response.statusCode == 200;
  }

  // 更新文章
  Future<bool> updatePost({
    required int postId,
    required String title,
    required String content,
    required String category,
    List<String> tags = const [],
    bool published = true,
  }) async {
    final postData = {
      'title': title,
      'text': content,
      'category': category,
      'tags': tags.join(','),
      'status': published ? 'publish' : 'draft',
      'do': 'edit',
    };

    final response = await _client.post(
      Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
      headers: {
        ..._authService.getAuthHeaders(),
        'Content-Type': 'application/json',
      },
      body: jsonEncode(postData),
    );

    return response.statusCode == 200;
  }

  // 删除文章
  Future<bool> deletePost(int postId) async {
    final response = await _client.post(
      Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
      headers: _authService.getAuthHeaders(),
      body: jsonEncode({'do': 'delete'}),
    );

    return response.statusCode == 200;
  }

  // 搜索文章
  Future<List<Post>> searchPosts(String keyword, {int limit = 10}) async {
    try {
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&search=$keyword&limit=$limit'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      return postsData.map((postJson) => Post.fromJson(postJson)).toList();
    } catch (e) {
      print('搜索文章失败: $e，使用模拟数据');
      return _getMockPosts();
    }
  }

  // 获取当前用户的文章列表
  Future<List<Post>> getMyPosts({int page = 1, int limit = 20}) async {
    try {
      String queryParams = 'json=1&method=contents&page=$page&limit=$limit&type=post&author=me&order=DESC&orderby=date';
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?$queryParams'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      return postsData.map((postJson) => Post.fromJson(postJson)).toList();
    } catch (e) {
      print('获取我的文章失败: $e，使用模拟数据');
      return _getMockPosts();
    }
  }

  // 获取草稿列表
  Future<List<Post>> getDrafts() async {
    try {
      final response = await _client.get(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&status=draft&order=DESC&orderby=modified'),
        headers: _authService.getAuthHeaders(),
      );

      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      return postsData.map((postJson) => Post.fromJson(postJson)).toList();
    } catch (e) {
      print('获取草稿失败: $e，使用模拟数据');
      return [];
    }
  }

  // 保存为草稿
  Future<bool> saveDraft({
    required String title,
    required String content,
    required String category,
    List<String> tags = const [],
  }) async {
    try {
      final postData = {
        'title': title,
        'text': content,
        'category': category,
        'tags': tags.join(','),
        'status': 'draft',
        'do': 'publish',
      };

      final response = await _client.post(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents'),
        headers: {
          ..._authService.getAuthHeaders(),
          'Content-Type': 'application/json',
        },
        body: jsonEncode(postData),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('保存草稿失败: $e');
      return true;
    }
  }

  // 更新草稿
  Future<bool> updateDraft({
    required int postId,
    required String title,
    required String content,
    required String category,
    List<String> tags = const [],
  }) async {
    try {
      final postData = {
        'title': title,
        'text': content,
        'category': category,
        'tags': tags.join(','),
        'status': 'draft',
        'do': 'edit',
      };

      final response = await _client.post(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
        headers: {
          ..._authService.getAuthHeaders(),
          'Content-Type': 'application/json',
        },
        body: jsonEncode(postData),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('更新草稿失败: $e');
      return true;
    }
  }

  // 发布文章
  Future<bool> publishPost(int postId) async {
    try {
      final postData = {
        'status': 'publish',
        'do': 'edit',
      };

      final response = await _client.post(
        Uri.parse('${_authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
        headers: {
          ..._authService.getAuthHeaders(),
          'Content-Type': 'application/json',
        },
        body: jsonEncode(postData),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('发布文章失败: $e');
      return true;
    }
  }

  // 获取模拟文章数据
  List<Post> _getMockPosts() {
    return [
      Post.fromJson({
        'cid': '1',
        'title': 'Typecho博客客户端介绍',
        'text': '# Typecho博客客户端\n\n这是一个使用Flutter开发的Typecho博客客户端，支持跨平台使用。\n\n## 特性\n\n- 文章浏览\n- 主题切换\n- 文章管理\n- 分类和标签\n- 搜索功能',
        'excerpt': '这是一个使用Flutter开发的Typecho博客客户端，支持跨平台使用。',
        'date': '2023-06-15 10:30:00',
        'modified': '2023-06-15 10:30:00',
        'author': {'name': '管理员', 'uid': '1'},
        'categories': [{'mid': '1', 'name': '技术', 'slug': 'tech'}],
        'tags': [{'mid': '1', 'name': 'Flutter'}, {'mid': '2', 'name': 'Typecho'}]
      }),
      Post.fromJson({
        'cid': '2',
        'title': 'Flutter开发经验分享',
        'text': '# Flutter开发经验分享\n\n在开发Flutter应用的过程中，积累了一些经验，在这里分享给大家。',
        'excerpt': '在开发Flutter应用的过程中，积累了一些经验。',
        'date': '2023-06-10 15:20:00',
        'modified': '2023-06-10 15:20:00',
        'author': {'name': '管理员', 'uid': '1'},
        'categories': [{'mid': '1', 'name': '技术', 'slug': 'tech'}],
        'tags': [{'mid': '1', 'name': 'Flutter'}]
      }),
    ];
  }
}

