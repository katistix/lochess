import type { Color } from '@lichess-org/chessground/types'
import { material } from '../lib/material'

// chessground draws pieces as <piece> elements; the icons here reuse them.
declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      piece: React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>
    }
  }
}

// The strip above or below the board, laid out as on chess.com: the name,
// and under it what this side has taken plus, for the side ahead, by how much.
// The pieces row keeps its height when empty so the board doesn't shift.
export function PlayerBar({ fen, color, label, active }: { fen: string; color: Color; label: string; active?: boolean }) {
  const { captured, advantage } = material(fen)[color]
  const taken: Color = color === 'white' ? 'black' : 'white'
  return (
    <div className="flex flex-col gap-0.5 py-1">
      <div className="flex items-center gap-2 text-sm">
        <span className="text-neutral-200">{label}</span>
        {active && <span className="size-2 rounded-full bg-emerald-400" />}
      </div>
      {/* cg-wrap lets chessground's piece images apply; it forces display: block, so the row is inside. */}
      <div className="cg-wrap captured">
        <div className="flex h-5 items-center gap-1.5">
          {captured.map(([role, n]) => (
            <span key={role} className="flex">
              {Array.from({ length: n }, (_, i) => (
                <piece key={i} className={`${role} ${taken}`} />
              ))}
            </span>
          ))}
          {advantage > 0 && <span className="text-xs text-neutral-400">+{advantage}</span>}
        </div>
      </div>
    </div>
  )
}
