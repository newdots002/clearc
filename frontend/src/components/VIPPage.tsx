import { useState, useEffect } from 'react'
import { Crown, Check, Zap, Shield, Clock, Sparkles, Loader2, ExternalLink, AlertCircle } from 'lucide-react'
import { GetVIPStatus, ActivateVIP } from '../../wailsjs/go/main/App'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

interface VIPStatus {
  isVip: boolean
  isTrialExpired: boolean
  trialDaysLeft: number
  trialDays: number
  firstUseTime: number
  vipActivatedAt: number
}

interface ActivationResult {
  success: boolean
  message: string
}

export default function VIPPage() {
  const [vipStatus, setVipStatus] = useState<VIPStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isActivating, setIsActivating] = useState(false)
  const [activationCode, setActivationCode] = useState('')
  const [showActivation, setShowActivation] = useState(false)
  const [activationMessage, setActivationMessage] = useState<{type: 'success' | 'error', text: string} | null>(null)

  useEffect(() => {
    loadVIPStatus()
  }, [])

  const loadVIPStatus = async () => {
    try {
      const status = await GetVIPStatus()
      setVipStatus(status)
    } catch (error) {
      console.error('Failed to load VIP status:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const formatActivationCode = (value: string) => {
    // 移除非字母数字字符，转大写
    const cleaned = value.replace(/[^A-Za-z0-9]/g, '').toUpperCase()
    // 每4个字符添加一个连字符
    const parts = cleaned.match(/.{1,4}/g) || []
    return parts.slice(0, 4).join('-')
  }

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatActivationCode(e.target.value)
    setActivationCode(formatted)
    setActivationMessage(null)
  }

  const handleActivate = async () => {
    if (!activationCode.trim() || activationCode.length !== 19) {
      setActivationMessage({ type: 'error', text: '请输入完整的激活码 (格式: XXXX-XXXX-XXXX-XXXX)' })
      return
    }
    
    setIsActivating(true)
    setActivationMessage(null)
    try {
      const result = await ActivateVIP(activationCode) as ActivationResult
      if (result.success) {
        setActivationMessage({ type: 'success', text: result.message })
        await loadVIPStatus()
        setShowActivation(false)
        setActivationCode('')
      } else {
        setActivationMessage({ type: 'error', text: result.message })
      }
    } catch (error) {
      console.error('Activation failed:', error)
      setActivationMessage({ type: 'error', text: '激活失败，请重试' })
    } finally {
      setIsActivating(false)
    }
  }

  const handleBuyClick = () => {
    // 打开购买页面
    BrowserOpenURL('https://clearc.top/pay.php')
  }

  const features = [
    {
      icon: Zap,
      title: '无限磁盘分析',
      description: '不限次数分析磁盘空间占用',
    },
    {
      icon: Shield,
      title: '智能清理建议',
      description: '获取专业的文件清理建议',
    },
    {
      icon: Clock,
      title: '定期后台分析',
      description: '自动后台分析，随时查看结果',
    },
    {
      icon: Sparkles,
      title: '持续更新',
      description: '永久享受所有新功能更新',
    },
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="w-8 h-8 animate-spin text-accent-blue" />
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="flex items-center justify-center gap-2">
          <Crown className="w-8 h-8 text-yellow-500" />
          <h1 className="font-primary text-3xl font-semibold text-text-primary">
            ClearC VIP
          </h1>
        </div>
        <p className="text-text-secondary">
          解锁全部功能，享受极致体验
        </p>
      </div>

      {/* Status Card */}
      <div className="max-w-2xl mx-auto">
        {vipStatus?.isVip ? (
          <div className="bg-gradient-to-r from-yellow-500/20 to-orange-500/20 border border-yellow-500/30 rounded-xl p-6 text-center space-y-3">
            <div className="flex items-center justify-center gap-2">
              <Crown className="w-6 h-6 text-yellow-500" />
              <span className="text-xl font-semibold text-yellow-500">永久 VIP 会员</span>
            </div>
            <p className="text-text-secondary text-sm">
              感谢您的支持！您已解锁所有功能
            </p>
            <p className="text-text-muted text-xs">
              激活时间：{new Date(vipStatus.vipActivatedAt * 1000).toLocaleDateString()}
            </p>
          </div>
        ) : (
          <div className="bg-bg-card border border-border rounded-xl p-6 space-y-4">
            <div className="text-center space-y-2">
              {vipStatus?.isTrialExpired ? (
                <>
                  <p className="text-accent-red font-medium">试用期已结束</p>
                  <p className="text-text-secondary text-sm">
                    购买永久 VIP 继续使用所有功能
                  </p>
                </>
              ) : (
                <>
                  <p className="text-accent-blue font-medium">
                    试用期剩余 {vipStatus?.trialDaysLeft} 天
                  </p>
                  <p className="text-text-secondary text-sm">
                    试用期内可免费使用所有功能
                  </p>
                </>
              )}
            </div>
            
            {/* Pricing */}
            <div className="bg-bg-hover rounded-xl p-6 text-center space-y-4">
              <div>
                <span className="text-4xl font-bold text-text-primary">¥5</span>
                <span className="text-text-muted ml-2">永久</span>
              </div>
              <p className="text-text-secondary text-sm">
                一次购买，终身使用
              </p>
              
              {!showActivation ? (
                <div className="space-y-3">
                  <button
                    onClick={handleBuyClick}
                    className="w-full py-3 bg-gradient-to-r from-yellow-500 to-orange-500 text-white font-medium rounded-lg hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
                  >
                    <ExternalLink size={16} />
                    立即购买
                  </button>
                  <button
                    onClick={() => setShowActivation(true)}
                    className="w-full py-2 bg-bg-card border border-border text-text-secondary rounded-lg hover:bg-border transition-colors text-sm"
                  >
                    已有激活码？点击激活
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  <p className="text-text-secondary text-xs">
                    请输入购买后获得的激活码
                  </p>
                  <input
                    type="text"
                    value={activationCode}
                    onChange={handleCodeChange}
                    placeholder="XXXX-XXXX-XXXX-XXXX"
                    maxLength={19}
                    className="w-full px-4 py-3 bg-bg-card border border-border rounded-lg text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue text-center font-mono text-lg tracking-wider"
                  />
                  {activationMessage && (
                    <div className={`flex items-center justify-center gap-2 text-sm ${
                      activationMessage.type === 'success' ? 'text-accent-green' : 'text-accent-red'
                    }`}>
                      {activationMessage.type === 'error' && <AlertCircle size={14} />}
                      {activationMessage.type === 'success' && <Check size={14} />}
                      {activationMessage.text}
                    </div>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={() => {
                        setShowActivation(false)
                        setActivationCode('')
                        setActivationMessage(null)
                      }}
                      className="flex-1 py-2 bg-bg-card border border-border text-text-primary rounded-lg hover:bg-border transition-colors"
                    >
                      取消
                    </button>
                    <button
                      onClick={handleActivate}
                      disabled={isActivating || activationCode.length !== 19}
                      className="flex-1 py-2 bg-accent-blue text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                    >
                      {isActivating ? (
                        <Loader2 size={16} className="animate-spin" />
                      ) : null}
                      激活 VIP
                    </button>
                  </div>
                  <p className="text-text-muted text-xs">
                    还没有激活码？<button onClick={handleBuyClick} className="text-accent-blue hover:underline">点击购买</button>
                  </p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Features */}
      <div className="max-w-2xl mx-auto">
        <h2 className="font-primary text-lg font-semibold text-text-primary mb-4 text-center">
          VIP 专属功能
        </h2>
        <div className="grid grid-cols-2 gap-4">
          {features.map((feature, index) => {
            const Icon = feature.icon
            return (
              <div
                key={index}
                className="bg-bg-card border border-border rounded-xl p-4 space-y-2"
              >
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 bg-accent-blue/20 rounded-lg flex items-center justify-center">
                    <Icon size={16} className="text-accent-blue" />
                  </div>
                  <span className="font-medium text-text-primary text-sm">
                    {feature.title}
                  </span>
                </div>
                <p className="text-text-muted text-xs">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>
      </div>

      {/* Benefits List */}
      <div className="max-w-2xl mx-auto bg-bg-card border border-border rounded-xl p-6">
        <h3 className="font-medium text-text-primary mb-4">购买后您将获得：</h3>
        <div className="space-y-3">
          {[
            '永久使用权，无需续费',
            '所有现有功能完整解锁',
            '未来新功能免费更新',
            '优先技术支持',
          ].map((benefit, index) => (
            <div key={index} className="flex items-center gap-2">
              <Check size={16} className="text-accent-green" />
              <span className="text-text-secondary text-sm">{benefit}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
