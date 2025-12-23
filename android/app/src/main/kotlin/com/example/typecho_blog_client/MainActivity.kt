package com.example.typecho_blog_client

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine

class MainActivity : FlutterActivity() {
    // Flutter embedding v2 不再需要覆盖onCreate方法来初始化Flutter
    // Flutter框架会自动处理Flutter引擎的初始化
    
    // 如果需要自定义Flutter引擎，可以覆盖configureFlutterEngine方法
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        // 在这里可以注册自定义通道或服务
        // GeneratedPluginRegistrant.registerWith(flutterEngine) // Flutter v2中不再需要这行代码
    }
}
