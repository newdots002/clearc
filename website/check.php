<?php
/**
 * ClearC 激活码验证接口
 * 供客户端验证激活码是否有效
 */

header('Content-Type: application/json; charset=utf-8');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: POST, GET, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type');

// 处理 OPTIONS 请求
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit;
}

// 激活码存储目录
define('CODES_DIR', __DIR__ . '/data/codes/');

/**
 * 返回 JSON 响应
 */
function jsonResponse($success, $message, $data = []) {
    echo json_encode(array_merge([
        'success' => $success,
        'message' => $message,
        'timestamp' => time()
    ], $data), JSON_UNESCAPED_UNICODE);
    exit;
}

/**
 * 验证激活码格式
 */
function validateCodeFormat($code) {
    // 格式: XXXX-XXXX-XXXX-XXXX
    return preg_match('/^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/', strtoupper($code));
}

/**
 * 获取激活码信息
 */
function getCodeInfo($code) {
    $code = strtoupper($code);
    $filename = CODES_DIR . $code . '.json';
    
    if (!file_exists($filename)) {
        return null;
    }
    
    return json_decode(file_get_contents($filename), true);
}

/**
 * 标记激活码已使用
 */
function markCodeUsed($code, $deviceId = null) {
    $code = strtoupper($code);
    $filename = CODES_DIR . $code . '.json';
    
    if (!file_exists($filename)) {
        return false;
    }
    
    $data = json_decode(file_get_contents($filename), true);
    $data['used'] = true;
    $data['used_at'] = date('Y-m-d H:i:s');
    $data['device_id'] = $deviceId;
    
    return file_put_contents($filename, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
}

// 获取请求参数
$action = $_GET['action'] ?? $_POST['action'] ?? 'verify';
$code = $_GET['code'] ?? $_POST['code'] ?? '';
$deviceId = $_GET['device_id'] ?? $_POST['device_id'] ?? '';

// 清理激活码格式
$code = strtoupper(trim($code));

// 检查激活码是否提供
if (empty($code)) {
    jsonResponse(false, '请提供激活码', ['error_code' => 'MISSING_CODE']);
}

// 验证激活码格式
if (!validateCodeFormat($code)) {
    jsonResponse(false, '激活码格式无效', ['error_code' => 'INVALID_FORMAT']);
}

// 根据操作类型处理
switch ($action) {
    case 'verify':
        // 仅验证激活码是否有效（不标记使用）
        $codeInfo = getCodeInfo($code);
        
        if ($codeInfo === null) {
            jsonResponse(false, '激活码不存在', ['error_code' => 'CODE_NOT_FOUND']);
        }
        
        if ($codeInfo['used']) {
            // 如果已使用，检查是否是同一设备
            if (!empty($deviceId) && $codeInfo['device_id'] === $deviceId) {
                jsonResponse(true, '激活码有效（已绑定当前设备）', [
                    'valid' => true,
                    'bound' => true,
                    'created_at' => $codeInfo['created_at']
                ]);
            } else {
                jsonResponse(false, '激活码已被其他设备使用', [
                    'error_code' => 'CODE_ALREADY_USED',
                    'used_at' => $codeInfo['used_at']
                ]);
            }
        }
        
        jsonResponse(true, '激活码有效', [
            'valid' => true,
            'bound' => false,
            'created_at' => $codeInfo['created_at']
        ]);
        break;
        
    case 'activate':
        // 激活（验证并标记使用）
        $codeInfo = getCodeInfo($code);
        
        if ($codeInfo === null) {
            jsonResponse(false, '激活码不存在', ['error_code' => 'CODE_NOT_FOUND']);
        }
        
        if ($codeInfo['used']) {
            // 如果已使用，检查是否是同一设备
            if (!empty($deviceId) && $codeInfo['device_id'] === $deviceId) {
                jsonResponse(true, '激活码已绑定当前设备', [
                    'activated' => true,
                    'already_bound' => true,
                    'created_at' => $codeInfo['created_at'],
                    'used_at' => $codeInfo['used_at']
                ]);
            } else {
                jsonResponse(false, '激活码已被其他设备使用', [
                    'error_code' => 'CODE_ALREADY_USED',
                    'used_at' => $codeInfo['used_at']
                ]);
            }
        }
        
        // 标记为已使用
        if (markCodeUsed($code, $deviceId)) {
            jsonResponse(true, '激活成功', [
                'activated' => true,
                'created_at' => $codeInfo['created_at'],
                'activated_at' => date('Y-m-d H:i:s')
            ]);
        } else {
            jsonResponse(false, '激活失败，请稍后重试', ['error_code' => 'ACTIVATION_FAILED']);
        }
        break;
        
    case 'status':
        // 查询激活码状态
        $codeInfo = getCodeInfo($code);
        
        if ($codeInfo === null) {
            jsonResponse(false, '激活码不存在', ['error_code' => 'CODE_NOT_FOUND']);
        }
        
        jsonResponse(true, '查询成功', [
            'code' => $code,
            'used' => $codeInfo['used'],
            'created_at' => $codeInfo['created_at'],
            'used_at' => $codeInfo['used_at'],
            'has_device' => !empty($codeInfo['device_id'])
        ]);
        break;
        
    default:
        jsonResponse(false, '未知操作', ['error_code' => 'UNKNOWN_ACTION']);
}
