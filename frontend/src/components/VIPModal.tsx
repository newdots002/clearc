import { useState } from 'react'
import { Crown, X, ExternalLink, Loader2, AlertCircle, Check } from 'lucide-react'
import { ActivateVIP } from '../../wailsjs/go/main/App'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

interface VIPModalProps {
  isOpen: boolean
  onClose: () => void
  onUpgrade: () => void
  trialDaysLeft?: number
}

interface ActivationResult {
  success: boolean
  message: string
}

export default function VIPModal({ isOpen, onClose, onUpgrade: _onUpgrade, trialDaysLeft }: VIPModalProps) {
  // Note: onUpgrade is kept for API compatibility but currently unused
  void _onUpgrade
  const [showActivation, setShowActivation] = useState(false)
  const [activationCode, setActivationCode] = useState('')
  const [isActivating, setIsActivating] = useState(false)
  const [activationMessage, setActivationMessage] = useState<{type: 'success' | 'error', text: string} | null>(null)

  if (!isOpen) return null

  const isExpired = trialDaysLeft !== undefined && trialDaysLeft <= 0

  const formatActivationCode = (value: string) => {
    const cleaned = value.replace(/[^A-Za-z0-9]/g, '').toUpperCase()
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
      setActivationMessage({ type: 'error', text: '请输入完整的激活码' })
      return
    }
    
    setIsActivating(true)
    setActivationMessage(null)
    try {
      const result = await ActivateVIP(activationCode) as ActivationResult
      if (result.success) {
        setActivationMessage({ type: 'success', text: result.message })
        setTimeout(() => {
          onClose()
          window.location.reload()
        }, 1500)
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
    BrowserOpenURL('https://clearc.top/pay.php?lang=zh')
  }

  const handleClose = () => {
    setShowActivation(false)
    setActivationCode('')
    setActivationMessage(null)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={handleClose}
      />
      
      {/* Modal */}
      <div className="relative bg-bg-card border border-border rounded-2xl p-6 w-full max-w-md mx-4 shadow-2xl">
        {/* Close button */}
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-text-muted hover:text-text-primary transition-colors"
        >
          <X size={20} />
        </button>

        {/* Content */}
        <div className="text-center space-y-4">
          {/* Icon */}
          <div className="flex justify-center">
            <div className="w-16 h-16 bg-gradient-to-br from-yellow-500 to-orange-500 rounded-2xl flex items-center justify-center">
              <Crown className="w-8 h-8 text-white" />
            </div>
          </div>

          {/* Title */}
          <div>
            <h2 className="text-xl font-semibold text-text-primary">
              {isExpired ? '试用期已结束' : '升级到永久 VIP'}
            </h2>
            <p className="text-text-secondary text-sm mt-1">
              {isExpired 
                ? '您的 10 天免费试用已结束'
                : `试用期还剩 ${trialDaysLeft} 天`
              }
            </p>
          </div>

          {/* Price */}
          <div className="bg-bg-hover rounded-xl p-4">
            <div className="flex items-baseline justify-center gap-1">
              <span className="text-3xl font-bold text-text-primary">¥5</span>
              <span className="text-text-muted">/ 永久</span>
            </div>
            <p className="text-text-muted text-xs mt-1">
              一次购买，终身使用所有功能
            </p>
          </div>

          {!showActivation ? (
            <>
              {/* Features */}
              <div className="text-left space-y-2 px-4">
                {[
                  '无限磁盘分析',
                  '智能清理建议',
                  '定期后台分析',
                  '持续功能更新',
                ].map((feature, index) => (
                  <div key={index} className="flex items-center gap-2 text-sm">
                    <div className="w-1.5 h-1.5 bg-accent-green rounded-full" />
                    <span className="text-text-secondary">{feature}</span>
                  </div>
                ))}
              </div>

              {/* Buttons */}
              <div className="space-y-2 pt-2">
                <button
                  onClick={handleBuyClick}
                  className="w-full py-3 bg-gradient-to-r from-yellow-500 to-orange-500 text-white font-medium rounded-xl hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
                >
                  <ExternalLink size={16} />
                  立即购买 VIP
                </button>
                <button
                  onClick={() => setShowActivation(true)}
                  className="w-full py-2 text-text-muted text-sm hover:text-text-secondary transition-colors"
                >
                  已有激活码？点击激活
                </button>
                <button
                  onClick={handleClose}
                  className="w-full py-2 text-text-muted text-sm hover:text-text-secondary transition-colors"
                >
                  {isExpired ? '稍后再说' : '继续试用'}
                </button>
              </div>
            </>
          ) : (
            <>
              {/* Activation Form */}
              <div className="space-y-3">
                <p className="text-text-secondary text-sm">
                  请输入购买后获得的激活码
                </p>
                <input
                  type="text"
                  value={activationCode}
                  onChange={handleCodeChange}
                  placeholder="XXXX-XXXX-XXXX-XXXX"
                  maxLength={19}
                  className="w-full px-4 py-3 bg-bg-hover border border-border rounded-lg text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue text-center font-mono text-lg tracking-wider"
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
              </div>

              {/* Buttons */}
              <div className="space-y-2 pt-2">
                <button
                  onClick={handleActivate}
                  disabled={isActivating || activationCode.length !== 19}
                  className="w-full py-3 bg-accent-blue text-white font-medium rounded-xl hover:bg-blue-600 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isActivating && <Loader2 size={16} className="animate-spin" />}
                  激活 VIP
                </button>
                <button
                  onClick={() => {
                    setShowActivation(false)
                    setActivationCode('')
                    setActivationMessage(null)
                  }}
                  className="w-full py-2 text-text-muted text-sm hover:text-text-secondary transition-colors"
                >
                  返回
                </button>
                <p className="text-text-muted text-xs">
                  还没有激活码？<button onClick={handleBuyClick} className="text-accent-blue hover:underline">点击购买</button>
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
