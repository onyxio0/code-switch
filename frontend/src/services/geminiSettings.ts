import { Call } from '@wailsio/runtime'

// 本地类型定义，避免依赖 CI 生成的绑定文件
export interface GeminiProxyStatus {
  enabled: boolean
  base_url: string
}

const serviceName = 'codeswitch/services.GeminiService'

export const fetchGeminiProxyStatus = async (): Promise<GeminiProxyStatus> => {
  return Call.ByName(`${serviceName}.ProxyStatus`)
}

export const enableGeminiProxy = async (): Promise<void> => {
  await Call.ByName(`${serviceName}.EnableProxy`)
}

export const disableGeminiProxy = async (): Promise<void> => {
  await Call.ByName(`${serviceName}.DisableProxy`)
}
