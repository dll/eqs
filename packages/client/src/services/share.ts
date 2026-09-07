/** 微信小程序分享能力封装；页面可直接复用返回值。 */
export interface ShareOptions {
  title: string
  path: string
  imageUrl?: string
}

export const enableShare = () => {
  // #ifdef MP-WEIXIN
  uni.showShareMenu({ withShareTicket: true })
  // #endif
}

export const createShareMessage = (options: ShareOptions) => ({
  title: options.title,
  path: options.path,
  ...(options.imageUrl ? { imageUrl: options.imageUrl } : {}),
})
