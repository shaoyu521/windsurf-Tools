import { main } from '../../wailsjs/go/models'

/** API Key / JWT / Refresh：首列凭证，其余为备注 */
export function parseCredentialLines(lines: string[]): Array<{ first: string; remark: string }> {
  return lines.map((line) => {
    const parts = line.trim().split(/\s+/)
    const first = parts[0] || ''
    const remark = parts.slice(1).join(' ').trim()
    return { first, remark }
  })
}

export function toAPIKeyItems(lines: string[]): main.APIKeyItem[] {
  return parseCredentialLines(lines).map(
    ({ first, remark }) => new main.APIKeyItem({ api_key: first, remark }),
  )
}

export function toJWTItems(lines: string[]): main.JWTItem[] {
  return parseCredentialLines(lines).map(({ first, remark }) => new main.JWTItem({ jwt: first, remark }))
}

export function toTokenItems(lines: string[]): main.TokenItem[] {
  return parseCredentialLines(lines).map(
    ({ first, remark }) => new main.TokenItem({ token: first, remark }),
  )
}

const EMAIL_RE = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+/i

function stripOuterQuotes(s: string): string {
  let t = s.trim()
  if (t.length >= 2 && t.startsWith('"') && t.endsWith('"')) {
    t = t.slice(1, -1).trim()
  }
  return t
}

function normalizeText(s: string): string {
  return s
    .replace(/\uFF1A/g, ':')
    .replace(/\uFF0C/g, ',')
    .replace(/\r/g, '')
}

/** 上一行仅有邮箱、下一行为「密码/密马/卡密」时合并为一行解析 */
export function mergePasswordContinuationLines(lines: string[]): string[] {
  const out: string[] = []
  for (let i = 0; i < lines.length; i++) {
    const cur = lines[i].trim()
    if (!cur) {
      continue
    }
    let j = i + 1
    while (j < lines.length && /^姓名\s*[:：]/i.test(lines[j].trim())) {
      j++
    }
    const nxt = lines[j]?.trim()
    const hasPwdInCur = /(?:密码|密\s*马|卡密\d*)\s*[:：]/i.test(cur)
    if (
      nxt &&
      EMAIL_RE.test(cur) &&
      !hasPwdInCur &&
      /^(?:卡密|密码|密\s*马)\s*[:：]/i.test(nxt) &&
      !EMAIL_RE.test(nxt) &&
      !/WFH-/i.test(nxt)
    ) {
      out.push(`${cur} ${nxt}`)
      i = j
    } else {
      out.push(cur)
    }
  }
  return out
}

function trimPwd(s: string): string {
  return s.replace(/^[,\s:：]+/, '').replace(/[，,;；\s]+$/g, '').trim()
}

/** 卡密行常见「卡密：密码:xxx」多一层前缀 */
function cleanPwd(s: string): string {
  let x = trimPwd(s)
  x = x.replace(/^密码\s*[:：]\s*/i, '').trim()
  return x
}

function item(
  email: string,
  password: string,
  remark: string,
  alt_password: string,
): main.EmailPasswordItem {
  return new main.EmailPasswordItem({
    email,
    password,
    remark: remark || '',
    alt_password: alt_password || '',
  })
}

/**
 * 解析单行邮箱+密码：JSON、账号:/邮箱:/卡号:、----、tab/逗号分隔、纯「邮箱 密码」等。
 */
export function parsePasswordLine(line: string): main.EmailPasswordItem | null {
  let s = normalizeText(stripOuterQuotes(line))
  if (!s) {
    return null
  }

  if (/^卡密\d*\s*[:：].*WFH-/i.test(s) || /^WFH-/i.test(s.trim())) {
    return null
  }

  let altPassword = ''
  const altM = s.match(/^(.+?)密码不对就这个(.+)$/i)
  if (altM) {
    s = altM[1].trim()
    altPassword = altM[2].trim()
  }

  if (s.startsWith('{')) {
    try {
      const o = JSON.parse(s) as {
        email?: string
        password?: string
        alt_password?: string
        remark?: string
      }
      if (o.email && o.password) {
        return item(
          o.email,
          o.password,
          o.remark || '',
          o.alt_password || altPassword,
        )
      }
    } catch {
      return null
    }
    return null
  }

  if (!EMAIL_RE.test(s)) {
    return null
  }

  let remark = ''

  if (/-{3,}/.test(s)) {
    const chunks = s.split(/-{3,}/).map((x) => x.trim()).filter(Boolean)
    if (chunks.length >= 2) {
      const em = chunks[0].match(EMAIL_RE)?.[0]
      if (!em) {
        return null
      }
      const pwdRaw = chunks.slice(1).join('---').trim()
      const pwd = cleanPwd(pwdRaw.replace(EMAIL_RE, '').trim() || pwdRaw)
      if (!pwd) {
        return null
      }
      return item(em, pwd, remark, altPassword)
    }
  }

  const emailMatch = s.match(EMAIL_RE)
  if (!emailMatch) {
    return null
  }
  const email = emailMatch[0]

  let after = s.slice(s.indexOf(email) + email.length).trim()
  after = after.replace(/^[,，\s:：]+/, '').trim()
  after = after.replace(/^(?:账号|邮箱|卡号\d*|邮件\s*号)\s*[:：]?\s*/i, '')

  const dateM = after.match(/\s+(\d{4}\/\d{1,2}\/\d{1,2})\s*$/)
  if (dateM) {
    remark = dateM[1]
    after = after.replace(/\s+\d{4}\/\d{1,2}\/\d{1,2}\s*$/, '').trim()
  }

  after = after.replace(/^(?:密码|密\s*马|卡密\d*)\s*[:：]\s*/i, '')
  after = cleanPwd(after)

  if (!after) {
    return null
  }

  if (after.includes('\t')) {
    const parts = after.split(/\t/).map((x) => x.trim()).filter(Boolean)
    const pwd =
      parts.find((p) => !/^\d{4}\/\d/.test(p) && p.length > 0) || parts[0] || ''
    const maybeDate = parts.find((p) => /^\d{4}\/\d/.test(p))
    if (maybeDate) {
      remark = remark || maybeDate
    }
    return item(email, cleanPwd(pwd), remark, altPassword)
  }

  return item(email, cleanPwd(after), remark, altPassword)
}

/**
 * 多行粘贴：合并续行、解析；同一邮箱多次出现时保留最后一次（避免重复导入）。
 */
export function toEmailPasswordItems(lines: string[]): main.EmailPasswordItem[] {
  const raw = lines.map((l) => l.trim()).filter(Boolean)
  const merged = mergePasswordContinuationLines(raw)
  const byEmail = new Map<string, main.EmailPasswordItem>()
  for (const line of merged) {
    const parsed = parsePasswordLine(line)
    if (parsed) {
      byEmail.set(parsed.email.toLowerCase(), parsed)
    }
  }
  return Array.from(byEmail.values())
}
