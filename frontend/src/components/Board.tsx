import { useEffect, useRef } from 'react'
import { Chessground } from '@lichess-org/chessground'
import type { Api } from '@lichess-org/chessground/api'
import type { Config } from '@lichess-org/chessground/config'
import type { Color, Key } from '@lichess-org/chessground/types'
import '@lichess-org/chessground/assets/chessground.base.css'
import '@lichess-org/chessground/assets/chessground.brown.css'
import '@lichess-org/chessground/assets/chessground.cburnett.css'
import type { game } from '../../wailsjs/go/models'

type Props = {
  state: game.State
  onMove: (from: Key, to: Key) => void
}

// Thin React wrapper around chessground. It only renders the state Go sends;
// legality comes from state.dests, never from the board itself.
export function Board({ state, onMove }: Props) {
  const el = useRef<HTMLDivElement>(null)
  const api = useRef<Api | null>(null)
  const onMoveRef = useRef(onMove)
  onMoveRef.current = onMove

  useEffect(() => {
    api.current = Chessground(el.current!, {
      premovable: { enabled: false },
      movable: {
        free: false,
        showDests: true,
        events: { after: (from, to) => onMoveRef.current(from, to) },
      },
    })
    return () => api.current?.destroy()
  }, [])

  useEffect(() => {
    api.current?.set(toConfig(state))
  }, [state])

  return <div ref={el} className="h-full w-full" />
}

function toConfig(s: game.State): Config {
  const turn = s.turn as Color
  const over = s.outcome !== '*'
  return {
    fen: s.fen,
    turnColor: turn,
    lastMove: s.lastMove as Key[],
    check: s.check ? turn : false,
    movable: {
      color: over ? undefined : turn,
      dests: new Map(Object.entries(s.dests) as [Key, Key[]][]),
    },
  }
}
