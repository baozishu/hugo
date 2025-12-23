class Post {
  final int id;
  final String title;
  final String slug;
  final String content;
  final String excerpt;
  final String date;
  final String author;
  final String category;
  final List<String> tags;

  Post({
    required this.id,
    required this.title,
    required this.slug,
    required this.content,
    required this.excerpt,
    required this.date,
    required this.author,
    required this.category,
    this.tags = const [],
  });

  factory Post.fromJson(Map<String, dynamic> json) {
    return Post(
      id: json['cid'],
      title: json['title'],
      slug: json['slug'],
      content: json['text'],
      excerpt: json['excerpt'] ?? json['text'].substring(0, json['text'].length < 100 ? json['text'].length : 100) + '...',
      date: json['date'],
      author: json['author'] ?? '未知作者',
      category: json['category'] ?? '未分类',
      tags: json['tags'] != null ? List<String>.from(json['tags']) : [],
    );
  }
}
