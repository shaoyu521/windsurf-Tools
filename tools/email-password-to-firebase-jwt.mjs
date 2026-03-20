/**
 * 批量用邮箱+密码调用 Firebase signIn，导出 idToken（JWT），供应用内「JWT」批量导入。
 *
 * 与密码导入写入的 acc.Token 一致，可显著减少导入时 UI 卡顿（JWT 导入不做 RegisterUser 等网络补全）。
 *
 * 用法:
 *   node tools/email-password-to-firebase-jwt.mjs tools/normalized-accounts-output.txt > jwt-import.txt
 *   node tools/email-password-to-firebase-jwt.mjs --concurrency=5 your.txt
 *   node tools/email-password-to-firebase-jwt.mjs accounts.txt --write-refresh=refresh-import.txt > jwt-import.txt
 *
 * 输入行格式: 账号: email 密码: pwd [ 2026/x/x ] [ 密码不对就这个 alt ]
 *
 * 输出: 每行「idToken 备注」，与前端 JWT 导入一致（首列为 JWT，空格后为备注）。
 */
import { readFileSync, writeFileSync } from 'node:fs'

/** 与 backend/services/windsurf.go FirebaseAPIKey 一致 */
const FIREBASE_WEB_KEY = 'AIzaSyDsOl-1XpT5err0Tcnx8FFod1H8gVGIycY'

const SIGN_IN_URL = `https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=${FIREBASE_WEB_KEY}`

function parseArgs(argv) {
  let concurrency = 4
  let file = null
  let writeRefresh = null
  for (const a of argv) {
    if (a.startsWith('--concurrency=')) {
      concurrency = Math.max(1, Math.min(20, parseInt(a.split('=')[1], 10) || 4))
    } else if (a.startsWith('--write-refresh=')) {
      writeRefresh = a.slice('--write-refresh='.length) || null
    } else if (!a.startsWith('-')) {
      file = a
    }
  }
  return { concurrency, file, writeRefresh }
}

/** @returns {{ email: string, passwords: string[] } | null} */
function parseAccountPasswordLine(line) {
  const s = line.trim()
  if (!s || s.startsWith('#')) return null
  const m = s.match(/^账号:\s*(\S+@\S+)\s*密码:\s*(.+)$/i)
  if (!m) return null
  let rest = m[2].trim()
  let alt = ''
  const altM = rest.match(/\s+密码不对就这个\s*(.+)$/i)
  if (altM) {
    alt = altM[1].trim()
    rest = rest.slice(0, altM.index).trim()
  }
  rest = rest.replace(/\s+\d{4}\/\d{1,2}\/\d{1,2}\s*$/, '').trim()
  const passwords = [rest, alt].filter(Boolean)
  if (!passwords.length) return null
  return { email: m[1], passwords }
}

async function signInWithPassword(email, password) {
  const body = JSON.stringify({
    returnSecureToken: true,
    email,
    password,
    clientType: 'CLIENT_TYPE_WEB',
  })
  const resp = await fetch(SIGN_IN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  })
  const text = await resp.text()
  if (!resp.ok) {
    let msg = text.slice(0, 300)
    try {
      const j = JSON.parse(text)
      msg = j.error?.message || msg
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }
  const j = JSON.parse(text)
  if (!j.idToken) throw new Error('响应无 idToken')
  if (!j.refreshToken) throw new Error('响应无 refreshToken')
  return { idToken: j.idToken, refreshToken: j.refreshToken }
}

async function tryLogin(email, passwords) {
  let lastErr = null
  for (const pw of passwords) {
    try {
      const { idToken, refreshToken } = await signInWithPassword(email, pw)
      return {
        idToken,
        refreshToken,
        usedAlt: passwords.length > 1 && pw === passwords[1],
      }
    } catch (e) {
      lastErr = e
    }
  }
  throw lastErr || new Error('登录失败')
}

async function main() {
  const argv = process.argv.slice(2)
  const { concurrency, file, writeRefresh } = parseArgs(argv)
  const raw = file && file !== '-' ? readFileSync(file, 'utf8') : readFileSync(0, 'utf8')

  const rows = []
  for (const line of raw.split('\n')) {
    const p = parseAccountPasswordLine(line)
    if (p) rows.push(p)
  }

  if (!rows.length) {
    console.error('未解析到任何「账号: … 密码: …」行')
    process.exit(1)
  }

  /** @type {{ email: string, idToken: string, refreshToken: string, usedAlt: boolean }[]} */
  const ok = []
  /** @type {{ email: string, err: string }[]} */
  const fail = []

  let done = 0
  for (let i = 0; i < rows.length; i += concurrency) {
    const chunk = rows.slice(i, i + concurrency)
    await Promise.all(
      chunk.map(async (row) => {
        try {
          const { idToken, refreshToken, usedAlt } = await tryLogin(row.email, row.passwords)
          ok.push({ email: row.email, idToken, refreshToken, usedAlt })
        } catch (e) {
          fail.push({ email: row.email, err: String(e?.message || e) })
        } finally {
          done++
          console.error(`# 进度 ${done}/${rows.length}`)
        }
      }),
    )
  }

  ok.sort((a, b) => a.email.localeCompare(b.email, 'en'))

  for (const { idToken, email, usedAlt } of ok) {
    const remark = usedAlt ? `${email} (备用密码)` : email
    console.log(`${idToken} ${remark}`)
  }

  if (writeRefresh) {
    const lines = ok.map(({ refreshToken, email, usedAlt }) => {
      const remark = usedAlt ? `${email} (备用密码)` : email
      return `${refreshToken} ${remark}`
    })
    writeFileSync(writeRefresh, `${lines.join('\n')}\n`, 'utf8')
  }

  console.error(`\n# 成功 ${ok.length} / ${rows.length}，失败 ${fail.length}`)
  if (writeRefresh) {
    console.error(
      `# 已写入 Refresh Token 行到 ${writeRefresh}（可留作续期；走「Refresh Token」导入仍会逐账号补全资料，可能较慢）`,
    )
  }
  for (const f of fail) {
    console.error(`# FAIL ${f.email}: ${f.err}`)
  }
}

main().catch((e) => {
  console.error(e)
  process.exit(1)
})
