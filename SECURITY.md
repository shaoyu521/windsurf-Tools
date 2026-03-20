# 安全说明：勿向 GitHub 提交敏感信息

## 绝对不要提交的内容

以下内容**不得**进入 Git 历史或公开仓库：

| 类型 | 说明 |
|------|------|
| **账号与密码** | 任意 `邮箱 + 密码` 列表、从 Windsurf / Codeium 导出的明文凭证 |
| **Refresh Token / JWT** | 可长期或短期登录的令牌字符串 |
| **个人 API Key** | 例如 `sk-ws-...` 形态的 Windsurf / Codeium 用户密钥（与 Firebase Web Key 不同） |
| **本地数据文件** | 应用生成的 `accounts.json`（内含密码、Token、API Key） |
| **调试脚本中的硬编码凭证** | 曾写在 `_quick_key.py` 等文件里的真实邮箱/密码（应仅用环境变量） |
| **工具输入/输出** | `tools/` 下含账号信息的 `.txt` / `.csv` / 日志等（见 `.gitignore`） |

若误提交，请立即**轮换密码/吊销 Token**，并考虑使用 `git filter-repo` / BFG 清理历史（必要时新建仓库）。

## 仓库中仍可能出现的「非用户密钥」

- `backend/services/windsurf.go` 中的 **Firebase Web API Key**（`FirebaseAPIKey`）与 Windsurf **网页端**客户端相同，用于 Firebase Identity Toolkit 的公开登录端点；它**不是**你的个人账号密码，但仍建议在 Google Cloud 控制台为该 Key 配置 **HTTP 引用来源等限制**。
- 桌面应用仅在**本机**将完整凭证写入用户配置目录下的 `accounts.json`，该路径**已被 `.gitignore` 排除在仓库之外**（仓库内不应出现该文件）。

## 工具脚本（`tools/`）

- `batch_quick_jwt.py`、`email-password-to-firebase-jwt.mjs`、`_quick_key.py` 需通过环境变量 **`FIREBASE_WEB_API_KEY`**（或 `QUICK_KEY_FIREBASE_WEB_API_KEY`）提供 Firebase Web Key，**不再**在脚本内写死，避免与「禁止上传密钥」政策混淆。
- 运行前请在**本机**设置环境变量；输入文件（含邮箱密码）仅留在本地，勿加入 Git。

### Windows（PowerShell）示例

```powershell
$env:FIREBASE_WEB_API_KEY = "<从 windsurf.go 或浏览器 Network 中 signIn 请求的 key= 复制>"
node tools/email-password-to-firebase-jwt.mjs .\your-local-accounts.txt
```

## 报告安全问题

若发现本仓库意外包含可复用的用户凭证或需私下沟通的安全问题，请通过 GitHub **Security Advisories** 或仓库维护者私下渠道联系，**勿**在公开 Issue 中粘贴密钥或密码。
