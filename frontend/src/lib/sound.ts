import type { settings } from '../../wailsjs/go/models'

// Kenney CC0 sounds (see assets/sounds/CREDITS.md). Like lichess, check uses the
// move sound and game start/end share one notification.
const files = import.meta.glob('../assets/sounds/*.ogg', { eager: true, query: '?url', import: 'default' }) as Record<
  string,
  string
>
const variants = (name: string) =>
  Object.entries(files)
    .filter(([path]) => path.includes(`/${name}`))
    .map(([, url]) => url)

const SOUNDS = {
  move: variants('move-'),
  capture: variants('capture-'),
  notify: variants('notify'),
}

export type Sound = keyof typeof SOUNDS

// Plays one of the sound's variations at random, respecting the user's settings.
export function playSound(kind: Sound, prefs: Pick<settings.Settings, 'sound' | 'volume'>) {
  if (!prefs.sound || prefs.volume <= 0) return
  const urls = SOUNDS[kind]
  const audio = new Audio(urls[Math.floor(Math.random() * urls.length)])
  audio.volume = prefs.volume / 100
  audio.play().catch(() => {}) // e.g. autoplay blocked before any interaction
}

// The sound for a move given its SAN: captures get their own sound.
export function moveSound(san: string): Sound {
  return san.includes('x') ? 'capture' : 'move'
}
