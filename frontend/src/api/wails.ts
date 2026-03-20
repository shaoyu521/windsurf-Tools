// 封装对 window.go 等后端的绑定调用
// Wails types map automatically to TS interfaces inside the auto-generated wailsjs folder generally.
// Here we define clean wrappers and types based on app.go

import * as AppHooks from '../../wailsjs/go/main/App';
import * as Models from '../../wailsjs/go/models';

export { AppHooks, Models };

// Specific typed helper types matching the Go struct
export interface ImportResult {
  email: string;
  success: boolean;
  error?: string;
}

export const APIInfo = {
  getAllAccounts: AppHooks.GetAllAccounts,
  deleteAccount: AppHooks.DeleteAccount,
  deleteExpiredAccounts: AppHooks.DeleteExpiredAccounts,
  deleteFreePlanAccounts: AppHooks.DeleteFreePlanAccounts,

  importByEmailPassword: AppHooks.ImportByEmailPassword,
  importByJWT: AppHooks.ImportByJWT,
  importByAPIKey: AppHooks.ImportByAPIKey,
  importByRefreshToken: AppHooks.ImportByRefreshToken,
  addSingleAccount: AppHooks.AddSingleAccount,

  switchAccount: AppHooks.SwitchAccount,
  autoSwitchToNext: AppHooks.AutoSwitchToNext,
  getCurrentWindsurfAuth: AppHooks.GetCurrentWindsurfAuth,
  getWindsurfAuthPath: AppHooks.GetWindsurfAuthPath,

  refreshAllTokens: AppHooks.RefreshAllTokens,
  refreshAllQuotas: AppHooks.RefreshAllQuotas,
  refreshAccountQuota: AppHooks.RefreshAccountQuota,

  getSettings: AppHooks.GetSettings,
  updateSettings: AppHooks.UpdateSettings,

  findWindsurfPath: AppHooks.FindWindsurfPath,
  applySeamlessPatch: AppHooks.ApplySeamlessPatch,
  restoreSeamlessPatch: AppHooks.RestoreSeamlessPatch,
  checkPatchStatus: AppHooks.CheckPatchStatus,
};
