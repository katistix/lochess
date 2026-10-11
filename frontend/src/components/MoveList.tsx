import { useEffect, useRef } from 'react'

export type MoveMark = { glyph: string; color: string }

type Props = {
  moves: string[] // SAN, white and black alternating
  outcome?: string // shown at the end when not "*"
  selected: number // index of the highlighted move; -1 for none
  marks?: (MoveMark | undefined)[] // optional annotation per move
  onSelect?: (index: number) => void
  className?: string
}

// Lichess-style move list: one row per move number.
export function MoveList({ moves, outcome, selected, marks, onSelect, className = 'h-64' }: Props) {
  const active = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    active.current?.scrollIntoView({ block: 'nearest' })
  }, [selected, moves.length])

  const rows: number[] = []
  for (let i = 0; i < moves.length; i += 2) rows.push(i)

  function cell(i: number) {
    if (i >= moves.length) return <span />
    const mark = marks?.[i]
    const isSel = i === selected
    return (
      <button
        ref={isSel ? active : undefined}
        onClick={onSelect ? () => onSelect(i) : undefined}
        className={`px-2 py-1 text-left ${onSelect ? 'cursor-pointer hover:bg-neutral-800' : 'cursor-default'} ${
          isSel ? 'bg-neutral-700 font-medium text-neutral-100' : 'text-neutral-300'
        }`}
      >
        {moves[i]}
        {mark && (
          <span className="ml-0.5 font-semibold" style={{ color: mark.color }}>
            {mark.glyph}
          </span>
        )}
      </button>
    )
  }

  return (
    <div className={`overflow-y-auto rounded bg-neutral-950/40 text-sm ${className}`}>
      {moves.length === 0 && <p className="p-3 text-neutral-600">No moves yet</p>}
      <div className="grid grid-cols-[2.5rem_1fr_1fr]">
        {rows.map((i) => (
          <div key={i} className="contents">
            <span className="bg-neutral-900/60 py-1 pr-2 text-right text-neutral-600">{i / 2 + 1}</span>
            {cell(i)}
            {cell(i + 1)}
          </div>
        ))}
      </div>
      {outcome && outcome !== '*' && (
        <p className="py-2 text-center font-medium text-neutral-300">{outcome.replace('1/2-1/2', '½-½')}</p>
      )}
    </div>
  )
}
