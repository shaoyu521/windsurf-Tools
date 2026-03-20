/**
 * 批量规范化邮箱+密码行，输出：账号: x 密码: y（含「密码不对就这个」时保留在同行以便再导入）
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'

const __dirname = dirname(fileURLToPath(import.meta.url))

const EMAIL_RE = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+/i

function stripOuterQuotes(s) {
  let t = s.trim()
  if (t.length >= 2 && t.startsWith('"') && t.endsWith('"')) {
    t = t.slice(1, -1).trim()
  }
  return t
}

function normalizeText(s) {
  return s.replace(/\uFF1A/g, ':').replace(/\uFF0C/g, ',').replace(/\r/g, '')
}

function trimPwd(s) {
  return s.replace(/^[,\s:：]+/, '').replace(/[，,;；\s]+$/g, '').trim()
}

/** 卡密行常见「卡密：密码:xxx」多一层前缀 */
function cleanPwd(p) {
  let x = trimPwd(p)
  x = x.replace(/^密码\s*[:：]\s*/i, '').trim()
  return x
}

function mergePasswordContinuationLines(lines) {
  const out = []
  for (let i = 0; i < lines.length; i++) {
    const cur = lines[i].trim()
    if (!cur) continue
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

function parseLine(line) {
  let s = normalizeText(stripOuterQuotes(line))
  if (!s) return null
  if (/^卡密\d*\s*[:：].*WFH-/i.test(s) || /^WFH-/i.test(s.trim())) return null

  let altPassword = ''
  const altM = s.match(/^(.+?)密码不对就这个(.+)$/i)
  if (altM) {
    s = altM[1].trim()
    altPassword = altM[2].trim()
  }

  if (!EMAIL_RE.test(s)) return null

  let remark = ''

  if (/-{3,}/.test(s)) {
    const chunks = s.split(/-{3,}/).map((x) => x.trim()).filter(Boolean)
    if (chunks.length >= 2) {
      const em = chunks[0].match(EMAIL_RE)?.[0]
      if (!em) return null
      const pwdRaw = chunks.slice(1).join('---').trim()
      const pwd = trimPwd(pwdRaw.replace(EMAIL_RE, '').trim() || pwdRaw)
      if (!pwd) return null
      return { email: em, password: cleanPwd(pwd), remark, altPassword }
    }
  }

  const emailMatch = s.match(EMAIL_RE)
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
  if (!after) return null

  if (after.includes('\t')) {
    const parts = after.split(/\t/).map((x) => x.trim()).filter(Boolean)
    const pwd =
      parts.find((p) => !/^\d{4}\/\d/.test(p) && p.length > 0) || parts[0] || ''
    const maybeDate = parts.find((p) => /^\d{4}\/\d/.test(p))
    if (maybeDate) remark = remark || maybeDate
    return { email, password: cleanPwd(pwd), remark, altPassword }
  }

  return { email, password: cleanPwd(after), remark, altPassword }
}

function formatRow({ email, password, remark, altPassword }) {
  let line = `账号: ${email} 密码: ${password}`
  if (remark) line += ` ${remark}`
  if (altPassword) line += ` 密码不对就这个${altPassword}`
  return line
}

const arg = process.argv[2]
let text
if (arg === '-' || arg === '--stdin') {
  text = readFileSync(0, 'utf8')
} else if (arg) {
  text = readFileSync(arg, 'utf8')
} else {
  const def = join(__dirname, 'raw-accounts-input.txt')
  text = readFileSync(def, 'utf8')
}
const lines = text.split('\n').map((l) => l.trim()).filter(Boolean)
const merged = mergePasswordContinuationLines(lines)
const byEmail = new Map()
for (const line of merged) {
  const p = parseLine(line)
  if (p) byEmail.set(p.email.toLowerCase(), p)
}

const sorted = Array.from(byEmail.values()).sort((a, b) =>
  a.email.localeCompare(b.email, 'en'),
)
for (const row of sorted) {
  console.log(formatRow(row))
}
console.error(`# stderr: 共 ${sorted.length} 条（按邮箱去重，后者覆盖前者）`)
