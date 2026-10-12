import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { GetSettings, SaveSettings } from '../../wailsjs/go/main/App'
import type { settings as gosettings } from '../../wailsjs/go/models'
import { DEFAULT_BOARD, DEFAULT_PIECES, themeCss } from './theme'

type Settings = gosettings.Settings

const Ctx = createContext<{ settings: Settings; update: (patch: Partial<Settings>) => void } | null>(null)

// Loads preferences from Go, applies the board/piece theme to the whole app,
// and saves every change immediately.
export function SettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<Settings | null>(null)

  useEffect(() => {
    GetSettings()
      .then(setSettings)
      .catch(() =>
        setSettings({
          board: DEFAULT_BOARD,
          pieces: DEFAULT_PIECES,
          sound: true,
          volume: 70,
          mode: 'bot',
          level: 1,
          color: 'white',
        }),
      )
  }, [])

  useEffect(() => {
    if (!settings) return
    let style = document.getElementById('lochess-theme')
    if (!style) {
      style = document.createElement('style')
      style.id = 'lochess-theme'
      document.head.appendChild(style)
    }
    style.textContent = themeCss(settings.board, settings.pieces)
  }, [settings])

  // Wait for saved settings so the board never flashes the default theme.
  if (!settings) return null

  function update(patch: Partial<Settings>) {
    const next = { ...settings!, ...patch }
    setSettings(next)
    SaveSettings(next).catch((e) => console.error('save settings:', e))
  }

  return <Ctx.Provider value={{ settings, update }}>{children}</Ctx.Provider>
}

export function useSettings() {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useSettings outside SettingsProvider')
  return ctx
}
