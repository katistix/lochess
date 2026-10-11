import type { Color } from '@lichess-org/chessground/types'
import { GRADES } from '../lib/grades'

// Round grade icon, as on chess.com/lichess review.
export function GradeBadge({ grade, className = 'h-5 w-5 text-[10px]' }: { grade: string; className?: string }) {
  const g = GRADES[grade]
  if (!g) return null
  return (
    <span
      title={g.label}
      className={`inline-flex shrink-0 items-center justify-center rounded-full font-bold leading-none text-white shadow ring-2 ring-black/20 ${className}`}
      style={{ backgroundColor: g.color }}
    >
      {g.icon}
    </span>
  )
}

// Positions a GradeBadge on the top-right corner of a square. Place inside a
// relatively positioned wrapper that has exactly the board's size.
export function SquareBadge({ square, grade, orientation }: { square: string; grade: string; orientation: Color }) {
  const file = square.charCodeAt(0) - 97 // a=0
  const rank = Number(square[1]) // 1-8
  const col = orientation === 'white' ? file : 7 - file
  const row = orientation === 'white' ? 8 - rank : rank - 1
  return (
    <div
      className="pointer-events-none absolute z-10 -translate-x-2/3 -translate-y-1/3"
      style={{ left: `${((col + 1) / 8) * 100}%`, top: `${(row / 8) * 100}%`, width: '5%', aspectRatio: '1' }}
    >
      <GradeBadge grade={grade} className="h-full w-full text-[min(1.6vh,1.2vw)]" />
    </div>
  )
}
