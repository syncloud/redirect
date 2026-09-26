import fs from 'node:fs'
import path from 'node:path'
import { SUPPORTED_LOCALES } from '../../src/i18n'

const dir = path.join(__dirname, '..', '..', 'src', 'locales')

function load (code) {
  return JSON.parse(fs.readFileSync(path.join(dir, code + '.json'), 'utf-8'))
}

function keys (messages, prefix = '') {
  return Object.entries(messages).flatMap(([key, value]) => {
    const full = prefix ? prefix + '.' + key : key
    return value !== null && typeof value === 'object' ? keys(value, full) : [full]
  })
}

const english = load('en')
const translated = SUPPORTED_LOCALES.map(l => l.code).filter(code => code !== 'en')

test.each(translated)('%s has exactly the english keys', code => {
  expect(keys(load(code)).sort()).toEqual(keys(english).sort())
})

test.each(SUPPORTED_LOCALES.map(l => l.code))('%s translates the price change notice', code => {
  const notice = load(code).account.priceChangeNotice
  expect(typeof notice).toBe('string')
  expect(notice).toContain('{date}')
  expect(notice).toContain('{monthly}')
  expect(notice).toContain('{annual}')
})

test.each(SUPPORTED_LOCALES.map(l => l.code))('%s takes the register price as a parameter', code => {
  const subtitle = load(code).register.subtitle
  expect(subtitle).toContain('{price}')
  expect(subtitle).not.toContain('£')
})
