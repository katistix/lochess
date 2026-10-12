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

// The strip above or below the board: who plays this side, what they have
// taken and, for the side ahead, by how much.
export function PlayerBar({ fen, color, label, active }: { fen: string; color: Color; label: string; active?: boolean }) {
  const { captured, advantage } = material(fen)[color]
  const taken: Color = color === 'white' ? 'black' : 'white'
  return (
    <div className="flex h-7 items-center gap-2 text-sm">
      <span className={`size-2 rounded-full ${active ? 'bg-emerald-400' : 'bg-transparent'}`} />
      <span className="text-neutral-200">{label}</span>
      <span className="cg-wrap captured flex items-center gap-1.5">
        {captured.map(([role, n]) => (
          <span key={role} className="flex">
            {Array.from({ length: n }, (_, i) => (
              <piece key={i} className={`${role} ${taken}`} />
            ))}
          </span>
        ))}
      </span>
      {advantage > 0 && <span className="text-xs text-neutral-400">+{advantage}</span>}
    </div>
  )
}
