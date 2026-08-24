const SCROLL_DELAY_MS = 300

function scrollIntoView (element) {
  if (element && element.tagName === 'INPUT') {
    element.scrollIntoView({ block: 'center', behavior: 'smooth' })
  }
}

export function startKeyboardTracking () {
  const viewport = window.visualViewport
  if (!viewport) {
    return () => {}
  }

  const onViewportChange = () => {
    const inset = Math.max(0, window.innerHeight - viewport.height - viewport.offsetTop)
    document.documentElement.style.setProperty('--kb', inset + 'px')
    if (inset > 0) {
      scrollIntoView(document.activeElement)
    }
  }

  const onFocusIn = (event) => {
    const target = event.target
    if (target && target.tagName === 'INPUT') {
      setTimeout(() => scrollIntoView(target), SCROLL_DELAY_MS)
    }
  }

  viewport.addEventListener('resize', onViewportChange)
  viewport.addEventListener('scroll', onViewportChange)
  document.addEventListener('focusin', onFocusIn)

  return () => {
    viewport.removeEventListener('resize', onViewportChange)
    viewport.removeEventListener('scroll', onViewportChange)
    document.removeEventListener('focusin', onFocusIn)
    document.documentElement.style.removeProperty('--kb')
  }
}
