import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:http_interceptor/http_interceptor.dart';

// 创建一个统一的HTTP客户端
class ApiClient {
  static final Client _client = InterceptedClient.build(
    interceptors: [RetryInterceptor(), LoggingInterceptor()],
    requestTimeout: Duration(seconds: 10),
  );

  static Client get client => _client;
}

// 请求重试拦截器
class RetryInterceptor implements InterceptorContract {
  final int maxRetries = 3;
  final Duration retryDelay = Duration(seconds: 1);

  @override
  FutureOr<RequestData> interceptRequest({required RequestData data}) async {
    return data;
  }

  @override
  FutureOr<ResponseData> interceptResponse({required ResponseData data}) async {
    // 如果请求成功或已达到最大重试次数，直接返回
    if (data.statusCode >= 200 && data.statusCode < 400 || 
        data.statusCode == 401 || // 认证错误不需要重试
        data.headers['X-Retry-Count'] != null && 
        int.parse(data.headers['X-Retry-Count']!) >= maxRetries) {
      return data;
    }

    // 递增重试计数
    final currentRetryCount = int.tryParse(data.headers['X-Retry-Count'] ?? '0') ?? 0;
    final nextRetryCount = currentRetryCount + 1;

    print('请求失败 (${data.statusCode}), 第 $nextRetryCount 次重试...');
    
    // 等待一段时间后重试
    await Future.delayed(retryDelay * nextRetryCount); // 指数退避策略
    
    // 添加重试计数到请求头
    final updatedHeaders = Map<String, String>.from(data.headers);
    updatedHeaders['X-Retry-Count'] = nextRetryCount.toString();
    
    // 创建新的请求数据并重新发送
    final newRequestData = RequestData(
      baseUrl: data.request?.baseUrl ?? '',
      path: data.request?.path ?? '',
      method: data.request?.method ?? '',
      headers: updatedHeaders,
      body: data.request?.body,
    );
    
    // 重新发送请求
    final client = http.Client();
    try {
      final uri = Uri.parse('${newRequestData.baseUrl}${newRequestData.path}');
      final request = http.Request(newRequestData.method, uri);
      request.headers.addAll(newRequestData.headers);
      if (newRequestData.body != null) {
        request.body = newRequestData.body!;
      }
      
      final response = await client.send(request);
      final responseBody = await response.stream.bytesToString();
      
      return ResponseData(
        statusCode: response.statusCode,
        body: responseBody,
        headers: Map.from(response.headers),
        request: newRequestData,
      );
    } finally {
      client.close();
    }
  }
}

// 请求日志拦截器
class LoggingInterceptor implements InterceptorContract {
  @override
  FutureOr<RequestData> interceptRequest({required RequestData data}) async {
    print('<--- API Request ---');
    print('URL: ${data.baseUrl}${data.path}');
    print('Method: ${data.method}');
    print('Headers: ${data.headers}');
    if (data.method == 'POST' || data.method == 'PUT') {
      print('Body: ${data.body}');
    }
    print('<------------------');
    return data;
  }

  @override
  FutureOr<ResponseData> interceptResponse({required ResponseData data}) async {
    print('<--- API Response ---');
    print('Status Code: ${data.statusCode}');
    print('Body: ${data.body}');
    print('<-------------------');
    
    // 统一处理错误
    if (data.statusCode >= 400) {
      final errorResponse = jsonDecode(data.body);
      throw ApiException(
        statusCode: data.statusCode,
        message: errorResponse['error'] ?? '请求失败',
      );
    }
    
    return data;
  }
}

// API异常类
class ApiException implements Exception {
  final int statusCode;
  final String message;

  ApiException({required this.statusCode, required this.message});

  @override
  String toString() {
    return 'API Error: $statusCode - $message';
  }
}

