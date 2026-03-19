# 认证相关

## 注册
POST /api/auth/signup
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

## 登录
POST /api/auth/signin
```json
{
  "identifier": "string",
  "password": "string"
}
```

## 登出
POST /api/auth/signout

## 刷新 Token
POST /api/auth/refresh 