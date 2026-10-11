import { createElement } from 'react'
import type { Color } from '@lichess-org/chessground/types'

const ROLES = [
  ['queen', 'q'],
  ['knight', 'n'],
  ['rook', 'r'],
  ['bishop', 'b'],
] as const

type Props = {
  square: string // destination, e.g. "e8"
  color: Color // side promoting
  orientation: Color
  onPick: (piece: 'q' | 'n' | 'r' | 'b') => void
  onCancel: () => void
}

// Lichess-style picker: the four pieces stacked on the promotion file,
// starting at the promotion square. Place inside the board's wrapper.
export function PromotionPicker({ square, color, orientation, onPick, onCancel }: Props) {
  const file = square.charCodeAt(0) - 97
  const rank = Number(square[1])
  const col = orientation === 'white' ? file : 7 - file
  const row = orientation === 'white' ? 8 - rank : rank - 1
  return (
    // cg-wrap gives the pieces chessground's piece images.
    <div className="cg-wrap !absolute inset-0 z-20 bg-black/50" onClick={onCancel}>
      {ROLES.map(([role, letter], i) => (
        <button
          key={role}
          title={role}
          onClick={(e) => {
            e.stopPropagation()
            onPick(letter)
          }}
          className="absolute rounded-full bg-neutral-300/90 hover:rounded-none hover:bg-orange-300"
          style={{ left: `${col * 12.5}%`, top: `${(row === 0 ? i : 7 - i) * 12.5}%`, width: '12.5%', height: '12.5%' }}
        >
          {createElement('piece', {
            className: `${role} ${color}`,
            style: { position: 'static', display: 'block', width: '100%', height: '100%', transform: 'none' },
          })}
        </button>
      ))}
    </div>
  )
}
