import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/services/auth_service.dart';

class ApiService {
  final AuthService authService;

  ApiService(this.authService);

  Future<List<Post>> getPosts({int page = 1, int limit = 10}) async {
    final response = await http.get(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=contents&page=$page&limit=$limit&type=post'),
      headers: authService.getAuthHeaders(),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final postsData = data['rows'] as List;
      return postsData.map((postJson) => Post.fromJson(postJson)).toList();
    } else {
      throw Exception('Failed to load posts');
    }
  }

  Future<Post> getPostDetail(int postId) async {
    final response = await http.get(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
      headers: authService.getAuthHeaders(),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return Post.fromJson(data);
    } else {
      throw Exception('Failed to load post detail');
    }
  }

  Future<List> getCategories() async {
    final response = await http.get(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=categories'),
      headers: authService.getAuthHeaders(),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return data['rows'] as List;
    } else {
      throw Exception('Failed to load categories');
    }
  }

  Future<bool> createPost(Map<String, dynamic> postData) async {
    final response = await http.post(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=contents'),
      headers: authService.getAuthHeaders(),
      body: jsonEncode(postData),
    );

    return response.statusCode == 200;
  }

  Future<bool> updatePost(int postId, Map<String, dynamic> postData) async {
    final response = await http.put(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
      headers: authService.getAuthHeaders(),
      body: jsonEncode(postData),
    );

    return response.statusCode == 200;
  }

  Future<bool> deletePost(int postId) async {
    final response = await http.delete(
      Uri.parse('${authService.apiUrl}/action/api?json=1&method=contents&cid=$postId'),
      headers: authService.getAuthHeaders(),
    );

    return response.statusCode == 200;
  }
}
