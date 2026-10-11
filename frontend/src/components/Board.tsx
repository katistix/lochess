import { useEffect, useRef } from 'react'
import { Chessground } from '@lichess-org/chessground'
import type { Api } from '@lichess-org/chessground/api'
import type { Config } from '@lichess-org/chessground/config'
import type { DrawShape } from '@lichess-org/chessground/draw'
import type { Color, Key } from '@lichess-org/chessground/types'
import '@lichess-org/chessground/assets/chessground.base.css'
import '@lichess-org/chessground/assets/chessground.brown.css'
import '@lichess-org/chessground/assets/chessground.cburnett.css'

export type BoardProps = {
  fen: string
  orientation: Color
  lastMove?: string[]
  check?: boolean
  dests?: Record<string, string[]> // legal moves; ignored unless movableColor is set
  movableColor?: Color // side the user may move; undefined makes the board read-only
  shapes?: DrawShape[] // arrows/circles drawn by the app
  onMove?: (from: Key, to: Key) => void
}

// Thin React wrapper around chessground. It only renders what Go decided;
// legality comes from dests, never from the board itself.
export function Board(props: BoardProps) {
  const el = useRef<HTMLDivElement>(null)
  const api = useRef<Api | null>(null)
  const onMoveRef = useRef(props.onMove)
  onMoveRef.current = props.onMove

  useEffect(() => {
    api.current = Chessground(el.current!, {
      premovable: { enabled: false },
      movable: {
        free: false,
        showDests: true,
        events: { after: (from, to) => onMoveRef.current?.(from, to) },
      },
    })
    return () => api.current?.destroy()
  }, [])

  useEffect(() => {
    api.current?.set(toConfig(props))
    api.current?.setAutoShapes(props.shapes ?? [])
  }, [props.fen, props.orientation, props.movableColor, props.dests, props.shapes, props.lastMove, props.check])

  return <div ref={el} className="h-full w-full" />
}

function toConfig(p: BoardProps): Config {
  const turn: Color = p.fen.split(' ')[1] === 'b' ? 'black' : 'white'
  return {
    fen: p.fen,
    orientation: p.orientation,
    turnColor: turn,
    lastMove: p.lastMove as Key[] | undefined,
    check: p.check ? turn : false,
    movable: {
      color: p.movableColor,
      dests: new Map(Object.entries(p.dests ?? {}) as [Key, Key[]][]),
    },
  }
}
