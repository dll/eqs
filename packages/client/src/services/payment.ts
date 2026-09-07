import { request } from '@/utils/request'

export interface PayParams {
  timeStamp: string
  nonceStr: string
  package: string
  signType?: 'RSA' | 'MD5'
  paySign: string
}

export interface PayResult {
  orderId: number
  paid: boolean
  cancelled?: boolean
  transaction?: any
}

/** 调起微信小程序收银台；H5/非微信端返回明确错误。 */
export const requestPayment = (params: PayParams): Promise<any> => new Promise((resolve, reject) => {
  // #ifdef MP-WEIXIN
  uni.requestPayment({
    provider: 'wxpay',
    timeStamp: params.timeStamp,
    nonceStr: params.nonceStr,
    package: params.package,
    signType: params.signType || 'RSA',
    paySign: params.paySign,
    success: resolve,
    fail: (error: any) => {
      const cancelled = String(error?.errMsg || '').toLowerCase().includes('cancel')
      reject({ cancelled, message: cancelled ? '已取消支付' : '支付失败，请重试', detail: error })
    },
  })
  // #ifndef MP-WEIXIN
  reject({ cancelled: false, message: '请在微信小程序中完成支付' })
  // #endif
})

/** 创建支付单。金额由服务端校验，客户端仅传订单及金额。 */
export async function payOrder(orderId: number, amount: number): Promise<PayResult> {
  const result: any = await request.post('/pay/create', { order_id: orderId, amount, channel: 'jsapi' })
  if (result.paid) return { orderId, paid: true, transaction: result.transaction }
  if (result.payParams) {
    try {
      await requestPayment(result.payParams as PayParams)
      return { orderId, paid: true, transaction: result.transaction }
    } catch (error: any) {
      if (error?.cancelled) return { orderId, paid: false, cancelled: true, transaction: result.transaction }
      throw error
    }
  }
  if (result.prepay_id) {
    // 后端当前返回 prepay_id；完整签名参数由后续接口/网关提供时再调起收银台。
    return { orderId, paid: false, transaction: result.transaction }
  }
  return { orderId, paid: false, transaction: result.transaction }
}
