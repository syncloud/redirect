import { ref } from 'vue'

const STORAGE_KEY = 'syncloud-theme'

function stored () {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved === 'dark' || saved === 'light' ? saved : null
  } catch {
    return null
  }
}

function systemPrefersDark () {
  return Boolean(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches)
}

export function preferredTheme () {
  return stored() || (systemPrefersDark() ? 'dark' : 'light')
}

function apply (value) {
  document.documentElement.setAttribute('data-theme', value)
}

export const theme = ref('light')

export function initTheme () {
  theme.value = preferredTheme()
  apply(theme.value)
}

export function setTheme (value) {
  theme.value = value === 'dark' ? 'dark' : 'light'
  try {
    localStorage.setItem(STORAGE_KEY, theme.value)
  } catch {
    // storage unavailable, theme still applies for this session
  }
  apply(theme.value)
}

export function toggleTheme () {
  setTheme(theme.value === 'dark' ? 'light' : 'dark')
}
