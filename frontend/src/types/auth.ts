export interface UserInfo {
  id: string
  username: string
}

export interface CaptchaResp {
  captchaId: string
  image: string
}

export interface LoginResp {
  accessToken: string
  expiresIn: number
  tokenType: string
  user: UserInfo
}

export interface RefreshResp {
  accessToken: string
  expiresIn: number
  tokenType: string
  user: UserInfo
}
