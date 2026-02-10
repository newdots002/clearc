<?php
/**
 * ClearC 支付页面
 * 处理易支付的支付请求和回调
 */

// 易支付配置 - 请修改为您的实际商户信息
define('EPAY_API_URL', 'https://pay.example.com/submit.php');  // 易支付网关地址
define('EPAY_PID', '1000');                                      // 商户ID
define('EPAY_KEY', 'your_secret_key_here');                      // 商户密钥

// 激活码存储目录（相对于网站根目录）
define('CODES_DIR', __DIR__ . '/data/codes/');

// 确保激活码目录存在
if (!file_exists(CODES_DIR)) {
    mkdir(CODES_DIR, 0755, true);
}

// 创建 .htaccess 禁止外网访问 data 目录
$htaccessPath = __DIR__ . '/data/.htaccess';
if (!file_exists($htaccessPath)) {
    file_put_contents($htaccessPath, "Order Deny,Allow\nDeny from all\n");
}

/**
 * 生成激活码
 */
function generateActivationCode() {
    $chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    $code = '';
    for ($i = 0; $i < 16; $i++) {
        if ($i > 0 && $i % 4 === 0) {
            $code .= '-';
        }
        $code .= $chars[random_int(0, strlen($chars) - 1)];
    }
    return $code;
}

/**
 * 保存激活码
 */
function saveActivationCode($code, $orderId, $tradeNo = '') {
    $data = [
        'code' => $code,
        'order_id' => $orderId,
        'trade_no' => $tradeNo,
        'created_at' => date('Y-m-d H:i:s'),
        'used' => false,
        'used_at' => null,
        'device_id' => null
    ];
    
    $filename = CODES_DIR . $code . '.json';
    return file_put_contents($filename, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
}

/**
 * 生成易支付签名
 */
function generateSign($params) {
    ksort($params);
    $signStr = '';
    foreach ($params as $key => $value) {
        if ($key !== 'sign' && $key !== 'sign_type' && $value !== '') {
            $signStr .= $key . '=' . $value . '&';
        }
    }
    $signStr = rtrim($signStr, '&') . EPAY_KEY;
    return md5($signStr);
}

/**
 * 验证易支付签名
 */
function verifySign($params) {
    $sign = $params['sign'] ?? '';
    return $sign === generateSign($params);
}

// 获取当前操作
$action = $_GET['action'] ?? $_POST['action'] ?? '';

// 处理异步通知 (notify_url)
if ($action === 'notify') {
    // 验证签名
    if (!verifySign($_GET)) {
        die('sign error');
    }
    
    $tradeStatus = $_GET['trade_status'] ?? '';
    $outTradeNo = $_GET['out_trade_no'] ?? '';
    $tradeNo = $_GET['trade_no'] ?? '';
    
    if ($tradeStatus === 'TRADE_SUCCESS') {
        // 生成激活码
        $activationCode = generateActivationCode();
        
        // 保存激活码
        saveActivationCode($activationCode, $outTradeNo, $tradeNo);
        
        // 记录订单与激活码的关联
        $orderFile = CODES_DIR . 'orders/' . $outTradeNo . '.json';
        if (!file_exists(dirname($orderFile))) {
            mkdir(dirname($orderFile), 0755, true);
        }
        file_put_contents($orderFile, json_encode([
            'order_id' => $outTradeNo,
            'trade_no' => $tradeNo,
            'activation_code' => $activationCode,
            'paid_at' => date('Y-m-d H:i:s')
        ], JSON_PRETTY_PRINT));
        
        echo 'success';
    } else {
        echo 'fail';
    }
    exit;
}

// 处理同步返回 (return_url)
if ($action === 'return') {
    $outTradeNo = $_GET['out_trade_no'] ?? '';
    $tradeNo = $_GET['trade_no'] ?? '';
    $tradeStatus = $_GET['trade_status'] ?? '';
    
    $paymentSuccess = false;
    $activationCode = '';
    
    // 验证签名
    if (verifySign($_GET) && $tradeStatus === 'TRADE_SUCCESS') {
        // 查找对应的激活码
        $orderFile = CODES_DIR . 'orders/' . $outTradeNo . '.json';
        if (file_exists($orderFile)) {
            $orderData = json_decode(file_get_contents($orderFile), true);
            $activationCode = $orderData['activation_code'] ?? '';
            $paymentSuccess = true;
        } else {
            // 如果订单文件不存在（可能notify还没处理完），生成新的激活码
            $activationCode = generateActivationCode();
            saveActivationCode($activationCode, $outTradeNo, $tradeNo);
            
            if (!file_exists(dirname($orderFile))) {
                mkdir(dirname($orderFile), 0755, true);
            }
            file_put_contents($orderFile, json_encode([
                'order_id' => $outTradeNo,
                'trade_no' => $tradeNo,
                'activation_code' => $activationCode,
                'paid_at' => date('Y-m-d H:i:s')
            ], JSON_PRETTY_PRINT));
            
            $paymentSuccess = true;
        }
    }
}

// 处理支付请求
if ($action === 'pay' && $_SERVER['REQUEST_METHOD'] === 'POST') {
    $paymentMethod = $_POST['method'] ?? 'alipay';
    
    // 生成订单号
    $orderId = 'CC' . time() . strtoupper(substr(md5(uniqid()), 0, 6));
    
    // 构建支付参数
    $params = [
        'pid' => EPAY_PID,
        'type' => $paymentMethod,
        'out_trade_no' => $orderId,
        'notify_url' => (isset($_SERVER['HTTPS']) ? 'https' : 'http') . '://' . $_SERVER['HTTP_HOST'] . '/pay.php?action=notify',
        'return_url' => (isset($_SERVER['HTTPS']) ? 'https' : 'http') . '://' . $_SERVER['HTTP_HOST'] . '/pay.php?action=return',
        'name' => 'ClearC永久VIP授权',
        'money' => '5.00',
        'sitename' => 'ClearC'
    ];
    
    // 生成签名
    $params['sign'] = generateSign($params);
    $params['sign_type'] = 'MD5';
    
    // 跳转到支付页面
    $payUrl = EPAY_API_URL . '?' . http_build_query($params);
    header('Location: ' . $payUrl);
    exit;
}

// 显示页面的变量
$showSuccess = isset($paymentSuccess) && $paymentSuccess;
$displayCode = $activationCode ?? '';
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>购买永久 VIP - ClearC</title>
    <meta name="description" content="购买 ClearC 永久 VIP 授权，一次付款，终身使用。">
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><rect fill='%2322D3EE' rx='15' width='100' height='100'/><text x='50' y='70' font-size='60' font-weight='bold' text-anchor='middle' fill='%230A0F1C'>C</text></svg>">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@500;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="style.css">
    <style>
        .pay-page {
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }
        
        .pay-content {
            flex: 1;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 120px 20px 60px;
        }
        
        .pay-container {
            max-width: 500px;
            width: 100%;
        }
        
        .pay-card {
            background: var(--bg-surface);
            border-radius: var(--radius-xl);
            padding: 40px;
            text-align: center;
        }
        
        .pay-icon {
            width: 80px;
            height: 80px;
            background: var(--accent-primary);
            border-radius: var(--radius-lg);
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
            font-size: 40px;
            font-weight: 700;
            color: var(--text-inverted);
        }
        
        .pay-title {
            font-size: 28px;
            font-weight: 700;
            color: var(--text-primary);
            margin-bottom: 8px;
        }
        
        .pay-subtitle {
            font-size: 16px;
            color: var(--text-tertiary);
            margin-bottom: 32px;
        }
        
        .pay-price-box {
            background: var(--bg-inset);
            border-radius: var(--radius-lg);
            padding: 24px;
            margin-bottom: 24px;
        }
        
        .pay-price {
            font-family: var(--font-mono);
            font-size: 48px;
            font-weight: 700;
            color: var(--accent-primary);
        }
        
        .pay-price-label {
            font-size: 14px;
            color: var(--text-tertiary);
            margin-top: 8px;
        }
        
        .pay-features {
            text-align: left;
            margin-bottom: 32px;
            list-style: none;
            padding: 0;
        }
        
        .pay-features li {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 12px 0;
            border-bottom: 1px solid var(--bg-inset);
            font-size: 14px;
            color: var(--text-secondary);
        }
        
        .pay-features li:last-child {
            border-bottom: none;
        }
        
        .pay-features .check {
            color: var(--accent-primary);
            font-weight: bold;
        }
        
        .pay-methods {
            display: flex;
            gap: 12px;
            margin-bottom: 24px;
        }
        
        .pay-method {
            flex: 1;
            padding: 16px;
            background: var(--bg-inset);
            border: 2px solid transparent;
            border-radius: var(--radius-md);
            cursor: pointer;
            transition: all 0.2s ease;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
        }
        
        .pay-method:hover {
            border-color: var(--text-muted);
        }
        
        .pay-method.active {
            border-color: var(--accent-primary);
            background: rgba(34, 211, 238, 0.1);
        }
        
        .pay-method-icon {
            font-size: 24px;
        }
        
        .pay-method-name {
            font-size: 13px;
            font-weight: 500;
            color: var(--text-secondary);
        }
        
        .pay-btn {
            width: 100%;
            padding: 16px 32px;
            font-size: 16px;
            cursor: pointer;
        }
        
        .pay-note {
            margin-top: 16px;
            font-size: 12px;
            color: var(--text-muted);
        }
        
        /* Success State */
        .pay-success {
            display: none;
        }
        
        .pay-success.active {
            display: block;
        }
        
        .pay-form.hidden {
            display: none;
        }
        
        .success-icon {
            width: 80px;
            height: 80px;
            background: #10B981;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
            color: white;
            font-size: 40px;
        }
        
        .activation-code-box {
            background: var(--bg-inset);
            border-radius: var(--radius-lg);
            padding: 24px;
            margin: 24px 0;
        }
        
        .activation-code-label {
            font-size: 14px;
            color: var(--text-tertiary);
            margin-bottom: 12px;
        }
        
        .activation-code {
            font-family: var(--font-mono);
            font-size: 24px;
            font-weight: 700;
            color: var(--accent-primary);
            letter-spacing: 2px;
            word-break: break-all;
            user-select: all;
        }
        
        .copy-btn {
            margin-top: 16px;
            padding: 12px 24px;
            font-size: 14px;
            cursor: pointer;
        }
        
        .activation-instructions {
            text-align: left;
            background: var(--bg-inset);
            border-radius: var(--radius-lg);
            padding: 20px;
            margin-top: 24px;
        }
        
        .activation-instructions h4 {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 12px;
        }
        
        .activation-instructions ol {
            padding-left: 20px;
            font-size: 13px;
            color: var(--text-tertiary);
            line-height: 1.8;
        }
        
        .back-link {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            margin-top: 24px;
            font-size: 14px;
            color: var(--text-tertiary);
            transition: color 0.2s ease;
            text-decoration: none;
        }
        
        .back-link:hover {
            color: var(--text-primary);
        }
        
        /* Loading State */
        .loading-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(10, 15, 28, 0.9);
            z-index: 2000;
            align-items: center;
            justify-content: center;
            flex-direction: column;
            gap: 20px;
        }
        
        .loading-overlay.active {
            display: flex;
        }
        
        .loading-spinner {
            width: 48px;
            height: 48px;
            border: 3px solid var(--bg-surface);
            border-top-color: var(--accent-primary);
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }
        
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
        
        .loading-text {
            font-size: 16px;
            color: var(--text-secondary);
        }
        
        /* Toast */
        .toast {
            position: fixed;
            bottom: 20px;
            left: 50%;
            transform: translateX(-50%) translateY(100px);
            background: var(--bg-surface);
            color: var(--text-primary);
            padding: 16px 24px;
            border-radius: var(--radius-md);
            font-size: 14px;
            z-index: 3000;
            transition: transform 0.3s ease;
            border: 1px solid var(--accent-primary);
        }
        
        .toast.show {
            transform: translateX(-50%) translateY(0);
        }
    </style>
</head>
<body class="pay-page">
    <!-- Header -->
    <header class="header">
        <div class="header-content">
            <a href="zh.html" class="logo-section">
                <div class="logo-icon">C</div>
                <span class="logo-name">ClearC</span>
            </a>
            <nav class="nav-section">
                <a href="zh.html#features" class="nav-link">功能特性</a>
                <a href="zh.html#pricing" class="nav-link">价格</a>
                <a href="zh.html#faq" class="nav-link">文档</a>
            </nav>
            <div class="cta-section">
                <a href="zh.html" class="btn btn-secondary">返回首页</a>
            </div>
        </div>
    </header>

    <!-- Pay Content -->
    <main class="pay-content">
        <div class="pay-container">
            <div class="pay-card">
                <!-- Payment Form -->
                <form class="pay-form <?php echo $showSuccess ? 'hidden' : ''; ?>" id="payForm" method="POST" action="pay.php?action=pay">
                    <div class="pay-icon">VIP</div>
                    <h1 class="pay-title">永久 VIP 授权</h1>
                    <p class="pay-subtitle">一次付款，终身使用所有功能</p>
                    
                    <div class="pay-price-box">
                        <div class="pay-price">¥5.00</div>
                        <div class="pay-price-label">永久授权 · 无需续费</div>
                    </div>
                    
                    <ul class="pay-features">
                        <li><span class="check">✓</span> 无限次磁盘分析</li>
                        <li><span class="check">✓</span> 后台自动扫描</li>
                        <li><span class="check">✓</span> 优先技术支持</li>
                        <li><span class="check">✓</span> 所有未来更新</li>
                        <li><span class="check">✓</span> 跨平台永久使用</li>
                    </ul>
                    
                    <div class="pay-methods">
                        <div class="pay-method active" data-method="alipay" onclick="selectMethod(this, 'alipay')">
                            <span class="pay-method-icon">💳</span>
                            <span class="pay-method-name">支付宝</span>
                        </div>
                        <div class="pay-method" data-method="wxpay" onclick="selectMethod(this, 'wxpay')">
                            <span class="pay-method-icon">💬</span>
                            <span class="pay-method-name">微信支付</span>
                        </div>
                        <div class="pay-method" data-method="qqpay" onclick="selectMethod(this, 'qqpay')">
                            <span class="pay-method-icon">🐧</span>
                            <span class="pay-method-name">QQ钱包</span>
                        </div>
                    </div>
                    
                    <input type="hidden" name="method" id="payMethod" value="alipay">
                    
                    <button type="submit" class="btn btn-primary pay-btn" id="payBtn">
                        立即支付 ¥5.00
                    </button>
                    
                    <p class="pay-note">支付完成后将自动生成激活码，请妥善保存</p>
                </form>
                
                <!-- Success State -->
                <div class="pay-success <?php echo $showSuccess ? 'active' : ''; ?>" id="paySuccess">
                    <div class="success-icon">✓</div>
                    <h1 class="pay-title">支付成功！</h1>
                    <p class="pay-subtitle">感谢您购买 ClearC 永久 VIP</p>
                    
                    <div class="activation-code-box">
                        <div class="activation-code-label">您的激活码</div>
                        <div class="activation-code" id="activationCode"><?php echo htmlspecialchars($displayCode); ?></div>
                        <button type="button" class="btn btn-secondary copy-btn" id="copyBtn">复制激活码</button>
                    </div>
                    
                    <div class="activation-instructions">
                        <h4>如何激活？</h4>
                        <ol>
                            <li>打开 ClearC 软件</li>
                            <li>点击左侧菜单的 "VIP 会员"</li>
                            <li>点击 "激活 VIP" 按钮</li>
                            <li>输入上方的激活码</li>
                            <li>点击确认，即可永久激活</li>
                        </ol>
                    </div>
                    
                    <a href="zh.html" class="back-link">
                        ← 返回首页
                    </a>
                </div>
            </div>
        </div>
    </main>

    <!-- Loading Overlay -->
    <div class="loading-overlay" id="loadingOverlay">
        <div class="loading-spinner"></div>
        <div class="loading-text">正在跳转支付页面...</div>
    </div>

    <!-- Toast -->
    <div class="toast" id="toast">已复制到剪贴板</div>

    <script>
        // 支付方式选择
        function selectMethod(el, method) {
            document.querySelectorAll('.pay-method').forEach(m => m.classList.remove('active'));
            el.classList.add('active');
            document.getElementById('payMethod').value = method;
        }
        
        // 表单提交时显示加载
        document.getElementById('payForm').addEventListener('submit', function() {
            document.getElementById('loadingOverlay').classList.add('active');
        });
        
        // 复制按钮
        document.getElementById('copyBtn').addEventListener('click', function() {
            const code = document.getElementById('activationCode').textContent;
            navigator.clipboard.writeText(code).then(() => {
                showToast('已复制到剪贴板');
            }).catch(() => {
                // Fallback
                const textArea = document.createElement('textarea');
                textArea.value = code;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                showToast('已复制到剪贴板');
            });
        });
        
        // Toast 显示
        function showToast(message) {
            const toast = document.getElementById('toast');
            toast.textContent = message;
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
            }, 2000);
        }
    </script>
</body>
</html>
