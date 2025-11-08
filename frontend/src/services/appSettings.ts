import { Call } from '@wailsio/runtime'

export type AppSettings = {
  show_heatmap: boolean
}

const DEFAULT_SETTINGS: AppSettings = {
  show_heatmap: true,
}

export const fetchAppSettings = async (): Promise<AppSettings> => {
  const data = await Call.ByName('codeswitch/services.AppSettingsService.GetAppSettings')
  return data ?? DEFAULT_SETTINGS
}

export const saveAppSettings = async (settings: AppSettings): Promise<AppSettings> => {
  return Call.ByName('codeswitch/services.AppSettingsService.SaveAppSettings', settings)
}
