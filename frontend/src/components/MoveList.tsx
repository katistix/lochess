import { useEffect, useRef } from 'react'

type Props = {
  moves: string[] // SAN, white and black alternating
  outcome: string // "*" while ongoing
}

// Lichess-style move list: one row per move number, latest move highlighted.
export function MoveList({ moves, outcome }: Props) {
  const end = useRef<HTMLDivElement>(null)

  useEffect(() => {
    end.current?.scrollIntoView({ block: 'nearest' })
  }, [moves.length])

  const rows: [string, string | undefined][] = []
  for (let i = 0; i < moves.length; i += 2) rows.push([moves[i], moves[i + 1]])
  const last = moves.length - 1

  return (
    <div className="h-64 overflow-y-auto rounded bg-neutral-950/40 text-sm">
      {rows.length === 0 && <p className="p-3 text-neutral-600">No moves yet</p>}
      <div className="grid grid-cols-[2.5rem_1fr_1fr]">
        {rows.map(([white, black], r) => (
          <Row key={r} n={r + 1} white={white} black={black} active={last - r * 2} />
        ))}
      </div>
      {outcome !== '*' && <p className="py-2 text-center font-medium text-neutral-300">{outcome}</p>}
      <div ref={end} />
    </div>
  )
}

// active is the index of the latest move relative to this row: 0 = white, 1 = black.
function Row({ n, white, black, active }: { n: number; white: string; black?: string; active: number }) {
  return (
    <>
      <span className="bg-neutral-900/60 py-1 pr-2 text-right text-neutral-600">{n}</span>
      <Cell san={white} active={active === 0} />
      <Cell san={black} active={active === 1} />
    </>
  )
}

function Cell({ san, active }: { san?: string; active: boolean }) {
  return (
    <span className={`px-2 py-1 ${active ? 'bg-neutral-700 font-medium text-neutral-100' : 'text-neutral-300'}`}>
      {san}
    </span>
  )
}
