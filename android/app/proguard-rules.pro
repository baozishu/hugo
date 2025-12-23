# Flutter Wrapper
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.**  { *; }
-keep class io.flutter.util.**  { *; }
-keep class io.flutter.view.**  { *; }
-keep class io.flutter.**  { *; }
-keep class io.flutter.plugins.**  { *; }

# 保留基本类型的包装类
-keep class java.lang.Integer { *; }
-keep class java.lang.Boolean { *; }
-keep class java.lang.Byte { *; }
-keep class java.lang.Character { *; }
-keep class java.lang.Double { *; }
-keep class java.lang.Float { *; }
-keep class java.lang.Long { *; }
-keep class java.lang.Short { *; }

# 保留应用程序入口点
-keep class com.example.typecho_blog_client.MainActivity { *; }

# 保留反射相关的类
-keepattributes Signature
-keepattributes *Annotation*

# 移除调试信息
-assumenosideeffects class android.util.Log {
    public static *** d(...);
    public static *** v(...);
    public static *** i(...);
    public static *** w(...);
    public static *** e(...);
}