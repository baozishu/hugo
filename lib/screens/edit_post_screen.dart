import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:typecho_blog_client/services/auth_service.dart';
import 'package:typecho_blog_client/services/typecho_api_service.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/components/custom_button.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';

class EditPostScreen extends StatefulWidget {
  final String? postId; // 如果是编辑，传入文章ID；如果是创建，则为null

  EditPostScreen({this.postId});

  @override
  _EditPostScreenState createState() => _EditPostScreenState();
}

class _EditPostScreenState extends State> EditPostScreen> {
  TextEditingController _titleController = TextEditingController();
  TextEditingController _contentController = TextEditingController();
  TextEditingController _excerptController = TextEditingController();
  TextEditingController _tagsController = TextEditingController();
  bool _isLoading = false;
  bool _isSaving = false;
  String? _errorMessage;
  Map<String, dynamic>? _selectedCategory;
  List<Map<String, dynamic>> _categories = [];

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _titleController.dispose();
    _contentController.dispose();
    _excerptController.dispose();
    _tagsController.dispose();
    super.dispose();
  }

  Future>void> _loadData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of>AuthService>(context, listen: false));
      
      // 加载分类列表
      _categories = await apiService.getCategories();
      
      // 如果是编辑文章，加载文章数据
      if (widget.postId != null) {
        final post = await apiService.getPostDetail(widget.postId!);
        _titleController.text = post.title;
        _contentController.text = post.content;
        _excerptController.text = post.excerpt ?? '';
        
        // 处理标签
        if (post.tags != null && post.tags!.isNotEmpty) {
          _tagsController.text = post.tags!.map((tag) => tag['name']).join(', ');
        }
        
        // 处理分类
        if (post.categories != null && post.categories!.isNotEmpty) {
          final categoryId = post.categories![0]['mid'];
          _selectedCategory = _categories.firstWhere(
            (cat) => cat['mid'] == categoryId,
            orElse: () => _categories.isNotEmpty ? _categories[0] : {},
          );
        }
      } else if (_categories.isNotEmpty) {
        // 默认选择第一个分类
        _selectedCategory = _categories[0];
      }
    } catch (e) {
      setState(() {
        _errorMessage = '加载数据失败: $e';
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future>void> _savePost() async {
    // 表单验证
    if (_titleController.text.trim().isEmpty) {
      _showError('标题不能为空');
      return;
    }
    
    if (_contentController.text.trim().isEmpty) {
      _showError('内容不能为空');
      return;
    }

    setState(() {
      _isSaving = true;
      _errorMessage = null;
    });

    try {
      final apiService = TypechoApiService(Provider.of>AuthService>(context, listen: false));
      
      // 构建文章数据
      final postData = {
        'title': _titleController.text.trim(),
        'text': _contentController.text.trim(),
        'excerpt': _excerptController.text.trim(),
        'category': _selectedCategory?['slug'] ?? '',
        'tags': _tagsController.text.trim().split(',').map((tag) => tag.trim()).toList(),
      };

      if (widget.postId != null) {
        // 更新文章
        await apiService.updatePost(widget.postId!, postData);
      } else {
        // 创建新文章
        await apiService.createPost(postData);
      }

      // 保存成功，返回上一页
      Navigator.pop(context, true);
    } catch (e) {
      setState(() {
        _errorMessage = '保存失败: $e';
      });
    } finally {
      setState(() {
        _isSaving = false;
      });
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.postId != null ? '编辑文章' : '创建文章'),
        actions: [
          CustomButton(
            onPressed: _isSaving ? null : _savePost,
            text: '保存',
            isLoading: _isSaving,
            icon: Icon(Icons.save),
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
                        onPressed: _loadData,
                        child: Text('重试'),
                      ),
                    ],
                  ),
                )
              : SingleChildScrollView(
                  padding: EdgeInsets.all(AppTheme.smallPadding),
                  child: Column(
                    children: [
                      // 标题输入框
                      TextField(
                        controller: _titleController,
                        decoration: InputDecoration(
                          labelText: '标题',
                          hintText: '请输入文章标题',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 1,
                      ),
                      SizedBox(height: AppTheme.smallPadding),

                      // 分类选择
                      _categories.isNotEmpty
                          ? DropdownButtonFormField(
                              value: _selectedCategory,
                              decoration: InputDecoration(
                                labelText: '分类',
                                border: OutlineInputBorder(),
                              ),
                              onChanged: (value) {
                                setState(() {
                                  _selectedCategory = value;
                                });
                              },
                              items: _categories.map((category) {
                                return DropdownMenuItem(
                                  value: category,
                                  child: Text(category['name']),
                                );
                              }).toList(),
                            )
                          : Container(),
                      SizedBox(height: AppTheme.smallPadding),

                      // 摘要输入框
                      TextField(
                        controller: _excerptController,
                        decoration: InputDecoration(
                          labelText: '摘要',
                          hintText: '请输入文章摘要（可选）',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 3,
                      ),
                      SizedBox(height: AppTheme.smallPadding),

                      // 标签输入框
                      TextField(
                        controller: _tagsController,
                        decoration: InputDecoration(
                          labelText: '标签',
                          hintText: '多个标签用逗号分隔（可选）',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 1,
                      ),
                      SizedBox(height: AppTheme.smallPadding),

                      // 内容输入框
                      TextField(
                        controller: _contentController,
                        decoration: InputDecoration(
                          labelText: '内容',
                          hintText: '请输入文章内容',
                          border: OutlineInputBorder(),
                          alignLabelWithHint: true,
                        ),
                        maxLines: 10,
                        keyboardType: TextInputType.multiline,
                      ),
                      SizedBox(height: AppTheme.mediumPadding),
                    ],
                  ),
                ),
    );
  }
}
