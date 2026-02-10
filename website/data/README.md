# 数据目录安全配置

此目录包含敏感数据（激活码），**禁止外网直接访问**。

## Apache 配置

已包含 `.htaccess` 文件，自动禁止访问。

## Nginx 配置

请在 Nginx 配置中添加以下规则：

```nginx
location /data {
    deny all;
    return 403;
}
```

或者在 server 块中添加：

```nginx
location ~ ^/data/ {
    internal;
    deny all;
}
```

## 目录结构

```
data/
├── .htaccess          # Apache 访问控制
├── README.md          # 本说明文件
└── codes/             # 激活码存储
    ├── XXXX-XXXX-XXXX-XXXX.json  # 激活码文件
    └── orders/        # 订单关联
        └── CC1234567890ABCDEF.json  # 订单文件
```

## 激活码文件格式

```json
{
    "code": "XXXX-XXXX-XXXX-XXXX",
    "order_id": "CC1234567890ABCDEF",
    "trade_no": "易支付交易号",
    "created_at": "2026-02-10 22:00:00",
    "used": false,
    "used_at": null,
    "device_id": null
}
```
