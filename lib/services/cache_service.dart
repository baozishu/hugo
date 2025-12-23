import 'package:shared_preferences/shared_preferences.dart';
import 'dart:convert';
import 'package:typecho_blog_client/models/post.dart';

class CacheService {
  static const String KEY_POSTS = 'cached_posts';
  static const String KEY_CATEGORIES = 'cached_categories';
  static const String KEY_TAGS = 'cached_tags';
  static const String KEY_POST_DETAIL = 'cached_post_detail_';
  static const int CACHE_DURATION = 10 * 60 * 1000; // 10分钟缓存

  Future<SharedPreferences> _prefs = SharedPreferences.getInstance();

  // 缓存文章列表
  Future<void> cachePosts(List<Post> posts) async {
    try {
      final prefs = await _prefs;
      final cacheData = {
        'data': posts.map((post) => post.toJson()).toList(),
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      };
      await prefs.setString(KEY_POSTS, json.encode(cacheData));
    } catch (e) {
      print('缓存文章失败: $e');
    }
  }

  // 获取缓存的文章列表
  Future<List<Post>?> getCachedPosts() async {
    try {
      final prefs = await _prefs;
      final cachedData = prefs.getString(KEY_POSTS);
      
      if (cachedData == null) return null;
      
      final data = json.decode(cachedData);
      final timestamp = data['timestamp'] as int;
      final now = DateTime.now().millisecondsSinceEpoch;
      
      // 检查缓存是否过期
      if (now - timestamp > CACHE_DURATION) {
        return null;
      }
      
      return (data['data'] as List)
          .map((item) => Post.fromJson(item as Map<String, dynamic>))
          .toList();
    } catch (e) {
      print('读取缓存文章失败: $e');
      return null;
    }
  }

  // 缓存单篇文章详情
  Future<void> cachePostDetail(int postId, Post post) async {
    try {
      final prefs = await _prefs;
      final cacheData = {
        'data': post.toJson(),
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      };
      await prefs.setString('$KEY_POST_DETAIL$postId', json.encode(cacheData));
    } catch (e) {
      print('缓存文章详情失败: $e');
    }
  }

  // 获取缓存的文章详情
  Future<Post?> getCachedPostDetail(int postId) async {
    try {
      final prefs = await _prefs;
      final cachedData = prefs.getString('$KEY_POST_DETAIL$postId');
      
      if (cachedData == null) return null;
      
      final data = json.decode(cachedData);
      final timestamp = data['timestamp'] as int;
      final now = DateTime.now().millisecondsSinceEpoch;
      
      // 检查缓存是否过期
      if (now - timestamp > CACHE_DURATION) {
        return null;
      }
      
      return Post.fromJson(data['data'] as Map<String, dynamic>);
    } catch (e) {
      print('读取缓存文章详情失败: $e');
      return null;
    }
  }

  // 清除所有缓存
  Future<void> clearCache() async {
    try {
      final prefs = await _prefs;
      await prefs.remove(KEY_POSTS);
      await prefs.remove(KEY_CATEGORIES);
      await prefs.remove(KEY_TAGS);
      
      // 清除所有文章详情缓存
      final keys = prefs.getKeys();
      for (var key in keys) {
        if (key.startsWith(KEY_POST_DETAIL)) {
          await prefs.remove(key);
        }
      }
    } catch (e) {
      print('清除缓存失败: $e');
    }
  }
}
