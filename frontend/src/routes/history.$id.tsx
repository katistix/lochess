import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'
import type { DrawShape } from '@lichess-org/chessground/draw'
import type { Color, Key } from '@lichess-org/chessground/types'
import { AnalyzeGame, DeleteGame, GetGame } from '../../wailsjs/go/main/App'
import type { history } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Board } from '../components/Board'
import { ConfirmButton } from '../components/ConfirmButton'
import { GradeBadge, SquareBadge } from '../components/GradeBadge'
import { MoveList, type MoveMark } from '../components/MoveList'
import { errorText } from '../lib/errors'
import { useSettings } from '../lib/settings'
import { moveSound, playSound } from '../lib/sound'
import { formatDate, opponentLabel } from '../lib/games'
import { GRADES, GRADE_ORDER, moveLabel } from '../lib/grades'

export const Route = createFileRoute('/history/$id')({
  component: Review,
})

const START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

function Review() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const [game, setGame] = useState<history.Game | null>(null)
  const [error, setError] = useState('')
  const [progress, setProgress] = useState<[number, number] | null>(null)
  // ply = number of moves played on the shown position; 0 is the start.
  const [ply, setPly] = useState(0)

  useEffect(() => {
    const off = EventsOn('analysis:progress', (gid: string, done: number, total: number) => {
      if (gid === id) setProgress([done, total])
    })
    GetGame(id)
      .then((g) => {
        setGame(g)
        if (!g.analysis) return AnalyzeGame(id).then(setGame)
      })
      .catch((e) => setError(errorText(e)))
      .finally(() => setProgress(null))
    return off
  }, [id])

  const total = game?.plies.length ?? 0
  const { settings } = useSettings()
  const go = useCallback(
    (p: number) => {
      const next = Math.max(0, Math.min(total, p))
      // Stepping forward one move sounds like playing it, as on lichess.
      if (game && next === ply + 1) playSound(moveSound(game.plies[next - 1].san), settings)
      setPly(next)
    },
    [total, game, ply, settings],
  )

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const keys: Record<string, number> = { ArrowLeft: ply - 1, ArrowRight: ply + 1, Home: 0, End: total }
      if (e.key in keys) {
        e.preventDefault()
        go(keys[e.key])
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [ply, total, go])

  const analysis = game?.analysis
  const grade = ply > 0 ? analysis?.moves[ply - 1] : undefined

  // Green arrow: what the engine wanted instead of the move just played.
  const shapes = useMemo<DrawShape[]>(() => {
    if (!grade || grade.class === 'best' || !grade.bestUci) return []
    return [{ orig: grade.bestUci.slice(0, 2) as Key, dest: grade.bestUci.slice(2, 4) as Key, brush: 'green' }]
  }, [grade])

  const marks = useMemo<(MoveMark | undefined)[]>(
    () =>
      analysis?.moves.map((m) => {
        const g = GRADES[m.class]
        return g?.glyph ? { glyph: g.glyph, color: g.color } : undefined
      }) ?? [],
    [analysis],
  )

  if (error && !game) {
    return (
      <div className="p-8">
        <p className="text-neutral-300">{error}</p>
        <Link to="/history" className="mt-2 inline-block text-sm text-neutral-400 underline underline-offset-4">
          Back to history
        </Link>
      </div>
    )
  }
  if (!game) return null

  const p = ply > 0 ? game.plies[ply - 1] : undefined
  const orientation = (game.mode === 'bot' ? game.humanColor : 'white') as Color
  const evalNow = analysis?.evals[ply]

  function remove() {
    DeleteGame(id)
      .then(() => navigate({ to: '/history' }))
      .catch((e) => setError(errorText(e)))
  }

  return (
    <div className="flex h-full items-center justify-center gap-6 p-6">
      <EvalBar win={evalNow?.win} text={evalNow?.text} orientation={orientation} />
      <div className="relative aspect-square w-[min(80vh,58vw)]">
        {p && grade && <SquareBadge square={p.uci.slice(2, 4)} grade={grade.class} orientation={orientation} />}
        <Board
          fen={p?.fen ?? START_FEN}
          lastMove={p ? [p.uci.slice(0, 2), p.uci.slice(2, 4)] : undefined}
          check={p?.check}
          orientation={orientation}
          shapes={shapes}
        />
      </div>

      <aside className="flex h-[min(80vh,58vw)] w-72 flex-col gap-4">
        <div>
          <p className="text-xs uppercase tracking-wide text-neutral-500">
            {opponentLabel(game)} · {formatDate(game.playedAt)}
          </p>
          <p className="mt-1 text-neutral-200">
            {game.result.replace('1/2-1/2', '½-½')} · {game.method}
          </p>
        </div>

        <MoveCard game={game} ply={ply} progress={progress} error={error} />

        <div className="grid grid-cols-4 gap-1">
          {[
            ['⏮', 0, 'First (Home)'],
            ['◀', ply - 1, 'Previous (←)'],
            ['▶', ply + 1, 'Next (→)'],
            ['⏭', total, 'Last (End)'],
          ].map(([label, target, title]) => (
            <button
              key={title}
              title={title as string}
              onClick={() => go(target as number)}
              className="rounded bg-neutral-800 py-1.5 text-sm text-neutral-300 hover:bg-neutral-700"
            >
              {label}
            </button>
          ))}
        </div>

        <MoveList
          moves={game.plies.map((x) => x.san)}
          outcome={game.result}
          selected={ply - 1}
          marks={marks}
          onSelect={(i) => go(i + 1)}
          className="min-h-24 flex-1"
        />

        {analysis && <Summary game={game} onJump={go} />}

        <div className="self-start">
          <ConfirmButton
            label="Delete game"
            confirmLabel="Click again to delete"
            onConfirm={remove}
            className="text-xs text-neutral-600 hover:text-red-400"
            armedClassName="rounded bg-red-500/20 px-2 py-1 text-xs text-red-300"
          />
        </div>
      </aside>
    </div>
  )
}

function MoveCard({
  game,
  ply,
  progress,
  error,
}: {
  game: history.Game
  ply: number
  progress: [number, number] | null
  error: string
}) {
  const a = game.analysis
  if (!a) {
    if (error) return <Card><p className="text-sm text-red-400">{error}</p></Card>
    const [done, total] = progress ?? [0, game.plies.length + 1]
    return (
      <Card>
        <p className="text-sm text-neutral-400">Analysing… {done}/{total}</p>
        <div className="mt-2 h-1 rounded bg-neutral-800">
          <div className="h-1 rounded bg-neutral-300 transition-all" style={{ width: `${(done / total) * 100}%` }} />
        </div>
      </Card>
    )
  }
  if (ply === 0) {
    return (
      <Card>
        <p className="text-sm text-neutral-400">Starting position. Use ← → to step through the game.</p>
      </Card>
    )
  }
  const m = a.moves[ply - 1]
  const g = GRADES[m.class]
  const san = game.plies[ply - 1].san
  return (
    <Card>
      <p className="flex items-center gap-2 text-sm">
        <GradeBadge grade={m.class} />
        <span>
          <span className="font-medium text-neutral-100">{moveLabel(ply - 1, san)}</span>{' '}
          <span style={{ color: g.color }}>
            {m.class === 'best' ? 'is the best move' : `is ${article(g.label)} ${g.label.toLowerCase()}`}
          </span>
        </span>
      </p>
      {m.class !== 'best' && m.bestSan && (
        <p className="mt-1 text-sm text-neutral-400">
          Best was <span className="font-medium text-green-400">{m.bestSan}</span>
        </p>
      )}
      <p className="mt-1 text-xs text-neutral-500">
        Eval {a.evals[ply - 1].text} → {a.evals[ply].text}
      </p>
    </Card>
  )
}

function Summary({ game, onJump }: { game: history.Game; onJump: (ply: number) => void }) {
  const a = game.analysis!
  const counts = (color: 0 | 1) =>
    Object.fromEntries(GRADE_ORDER.map((c) => [c, a.moves.filter((m, i) => i % 2 === color && m.class === c).length]))
  const white = counts(0)
  const black = counts(1)
  const name = (c: string) => (game.mode === 'bot' ? (game.humanColor === c ? 'You' : 'Stockfish') : c === 'white' ? 'White' : 'Black')

  return (
    <div className="flex flex-col gap-3 text-sm">
      <table className="w-full">
        <thead>
          <tr className="text-xs text-neutral-500">
            <th />
            <th className="text-right font-normal">{name('white')}</th>
            <th className="text-right font-normal">{name('black')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="text-neutral-400">Accuracy</td>
            <td className="text-right font-medium">{Math.round(a.whiteAccuracy)}%</td>
            <td className="text-right font-medium">{Math.round(a.blackAccuracy)}%</td>
          </tr>
          {GRADE_ORDER.slice(3).map((c) => (
            <tr key={c}>
              <td style={{ color: GRADES[c].color }}>{GRADES[c].plural}</td>
              <td className="text-right text-neutral-300">{white[c]}</td>
              <td className="text-right text-neutral-300">{black[c]}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {a.keyMoments.length > 0 && (
        <div>
          <p className="mb-1 text-xs text-neutral-500">Key moments</p>
          <div className="flex flex-wrap gap-1">
            {a.keyMoments.map((i) => {
              const g = GRADES[a.moves[i].class]
              return (
                <button
                  key={i}
                  onClick={() => onJump(i + 1)}
                  className="rounded bg-neutral-800 px-2 py-1 text-xs hover:bg-neutral-700"
                >
                  {moveLabel(i, game.plies[i].san)}
                  <span style={{ color: g.color }}>{g.glyph}</span>
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

function EvalBar({ win, text, orientation }: { win?: number; text?: string; orientation: Color }) {
  // White's share of the bar, measured from the bottom when White is at the bottom.
  const white = win ?? 50
  const fromBottom = orientation === 'white' ? white : 100 - white
  const bottomColor = orientation === 'white' ? 'bg-neutral-100' : 'bg-neutral-900'
  const topColor = orientation === 'white' ? 'bg-neutral-900' : 'bg-neutral-100'
  // The number sits on the side that is ahead, in a colour readable on it.
  const whiteAhead = white >= 50
  return (
    <div className={`relative h-[min(80vh,58vw)] w-4 overflow-hidden rounded-sm ${topColor} ring-1 ring-neutral-700`}>
      <div className={`absolute bottom-0 w-full transition-all duration-300 ${bottomColor}`} style={{ height: `${fromBottom}%` }} />
      {text && (
        <span
          className={`absolute left-1/2 -translate-x-1/2 text-[9px] font-semibold ${
            whiteAhead === (orientation === 'white') ? 'bottom-1' : 'top-1'
          } ${whiteAhead ? 'text-neutral-900' : 'text-neutral-100'}`}
        >
          {text.replace(/^\+/, '')}
        </span>
      )}
    </div>
  )
}

function Card({ children }: { children: React.ReactNode }) {
  return <div className="rounded bg-neutral-800/60 p-3">{children}</div>
}

function article(word: string) {
  return /^[aeiou]/i.test(word) ? 'an' : 'a'
}
