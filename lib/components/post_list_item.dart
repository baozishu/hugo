import 'package:flutter/material.dart';
import 'package:typecho_blog_client/models/post.dart';
import 'package:typecho_blog_client/theme/app_theme.dart';
import 'package:cached_network_image/cached_network_image.dart';

class PostListItem extends StatelessWidget {
  final Post post;
  final Function() onTap;
  
  const PostListItem({Key? key, required this.post, required this.onTap}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: EdgeInsets.symmetric(horizontal: AppTheme.mediumPadding, vertical: 8),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppTheme.mediumRadius),
        ),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(AppTheme.mediumRadius),
          child: Padding(
            padding: AppTheme.mediumPadding,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 文章缩略图
                if (post.thumbnail.isNotEmpty)
                  Container(
                    margin: EdgeInsets.only(bottom: 12),
                    height: 160,
                    width: double.infinity,
                    child: CachedNetworkImage(
                      imageUrl: post.thumbnail,
                      imageBuilder: (context, imageProvider) => Container(
                        decoration:
                          BoxDecoration(
                          borderRadius: BorderRadius.circular(AppTheme.mediumRadius),
                          image: DecorationImage(
                            image: imageProvider,
                            fit: BoxFit.cover,
                          ),
                        ),
                      ),
                      placeholder: (context, url) => Container(
                        decoration:
                          BoxDecoration(
                          borderRadius: BorderRadius.circular(AppTheme.mediumRadius),
                          color: Colors.grey[200],
                        ),
                        child: Center(child: Icon(Icons.image_outlined, color: Colors.grey)),
                      ),
                      errorWidget: (context, url, error) => Container(
                        decoration:
                          BoxDecoration(
                          borderRadius: BorderRadius.circular(AppTheme.mediumRadius),
                          color: Colors.grey[200],
                        ),
                        child: Center(child: Icon(Icons.error_outline, color: Colors.red)),
                      ),
                      fit: BoxFit.cover,
                      // 设置缓存策略
                      cacheKey: post.thumbnail,
                    ),
                  ),
                
                // 文章标题
                Text(
                  post.title,
                  style: TextStyle(
                    fontSize: AppTheme.largeFontSize,
                    fontWeight: FontWeight.bold,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                SizedBox(height: 8),
                
                // 作者和日期
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      post.author,
                      style: TextStyle(
                        fontSize: AppTheme.smallFontSize,
                        color: Theme.of(context).textTheme.bodyText2?.color,
                      ),
                    ),
                    Text(
                      _formatDate(post.date),
                      style: TextStyle(
                        fontSize: AppTheme.smallFontSize,
                        color: Theme.of(context).textTheme.bodyText2?.color,
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 8),
                
                // 文章摘要
                Text(
                  post.excerpt,
                  style: TextStyle(
                    fontSize: AppTheme.normalFontSize,
                    color: Theme.of(context).textTheme.bodyText2?.color,
                  ),
                  maxLines: post.thumbnail.isEmpty ? 4 : 2, // 没有缩略图时显示更多摘要
                  overflow: TextOverflow.ellipsis,
                ),
                SizedBox(height: 12),
                
                // 分类和标签
                Row(
                  children: [
                    Container(
                      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                        color: Theme.of(context).primaryColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(AppTheme.smallRadius),
                      ),
                      child: Text(
                        post.category,
                        style: TextStyle(
                          fontSize: AppTheme.smallFontSize,
                          color: Theme.of(context).primaryColor,
                        ),
                      ),
                    ),
                    if (post.tags.isNotEmpty)
                      Expanded(
                        child: Padding(
                          padding: EdgeInsets.only(left: 8),
                          child: Wrap(
                            spacing: 6,
                            runSpacing: 4,
                            children: post.tags.take(3).map((tag) {
                              return Container(
                                padding: EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                decoration: BoxDecoration(
                                  color: Theme.of(context).accentColor.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(AppTheme.smallRadius),
                                ),
                                child: Text(
                                  tag,
                                  style: TextStyle(
                                    fontSize: AppTheme.smallFontSize,
                                    color: Theme.of(context).accentColor,
                                  ),
                                ),
                              );
                            }).toList(),
                          ),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  String _formatDate(String dateStr) {
    try {
      DateTime date = DateTime.parse(dateStr);
      return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
    } catch (e) {
      return dateStr;
    }
  }
}