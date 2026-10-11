import { useEffect, useRef, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import type { Color } from '@lichess-org/chessground/types'
import { BotMove, GameState, Move, NewGame, Resign, Setup } from '../../wailsjs/go/main/App'
import type { match } from '../../wailsjs/go/models'
import { Board } from '../components/Board'
import { ConfirmButton } from '../components/ConfirmButton'
import { MoveList } from '../components/MoveList'

export const Route = createFileRoute('/play')({
  component: Play,
})

type Mode = 'bot' | 'local'
type ColorChoice = 'white' | 'random' | 'black'

const LEVELS = [1, 2, 3, 4, 5, 6, 7, 8]

function Play() {
  const [state, setState] = useState<match.State | null>(null)
  const [error, setError] = useState('')
  const [mode, setMode] = useState<Mode>('bot')
  const [level, setLevel] = useState(1)
  const [color, setColor] = useState<ColorChoice>('white')

  // Bumped on every new game so replies from a previous game are dropped.
  const gen = useRef(0)
  function accept(g: number) {
    return (s: match.State) => {
      if (g === gen.current) setState(s)
    }
  }
  function resync(g: number) {
    return (err: unknown) => {
      setError(String(err))
      GameState().then(accept(g))
    }
  }

  useEffect(() => {
    GameState().then(setState)
  }, [])

  // Whenever it's the bot's turn, ask Go to play. Duplicate calls are no-ops in Go.
  useEffect(() => {
    if (!state?.botToMove) return
    const g = gen.current
    BotMove().then(accept(g)).catch(resync(g))
  }, [state])

  function startGame() {
    const g = ++gen.current
    setError('')
    NewGame({ mode, level, color }).then(accept(g)).catch(resync(g))
  }

  function handleMove(from: string, to: string) {
    setError('')
    // In setup, the first move as White starts the game with the chosen options.
    const g = state?.phase === 'setup' ? ++gen.current : gen.current
    const started = state?.phase === 'setup' ? NewGame({ mode, level, color }) : Promise.resolve()
    started
      .then(() => Move(from + to))
      .then(accept(g))
      .catch(resync(g))
  }

  function run(call: () => Promise<match.State>) {
    const g = ++gen.current
    setError('')
    call().then(accept(g)).catch(resync(g))
  }

  if (!state) return null

  const g = state.game
  const setup = state.phase === 'setup'
  // In setup the board previews the chosen side; White (or 1v1) can move to start.
  const canStartByMoving = mode === 'local' || color === 'white'
  let orientation: Color = 'white'
  if (setup && mode === 'bot' && color === 'black') orientation = 'black'
  if (!setup && state.mode === 'bot') orientation = state.humanColor as Color

  let movableColor: Color | undefined
  if (setup && canStartByMoving) movableColor = 'white'
  if (state.phase === 'playing' && !state.botToMove) movableColor = g.turn as Color

  return (
    <div className="flex h-full items-center justify-center gap-8 p-6">
      <div className="aspect-square w-[min(80vh,65vw)]">
        <Board
          fen={g.fen}
          lastMove={g.lastMove}
          check={g.check}
          dests={g.dests}
          orientation={orientation}
          movableColor={movableColor}
          onMove={handleMove}
        />
      </div>

      <aside className="flex w-64 flex-col gap-5">
        {setup ? (
          <SetupPanel
            mode={mode}
            setMode={setMode}
            level={level}
            setLevel={setLevel}
            color={color}
            setColor={setColor}
            hint={canStartByMoving ? 'Make a move to start, or' : undefined}
            onStart={startGame}
          />
        ) : (
          <>
            <div>
              <p className="text-xs uppercase tracking-wide text-neutral-500">
                {state.mode === 'bot' ? `Stockfish level ${state.level}` : 'Local game'}
              </p>
              <p className="mt-1 text-neutral-200">{status(state)}</p>
            </div>
            <MoveList moves={g.moves} outcome={g.outcome} selected={g.moves.length - 1} />
            {state.phase === 'playing' ? (
              <div className="self-start">
                <ConfirmButton
                  label="Resign"
                  confirmLabel="Click again to resign"
                  onConfirm={() => run(Resign)}
                  className="text-sm text-neutral-500 hover:text-neutral-200"
                  armedClassName="rounded bg-red-500/20 px-2 py-1 text-sm text-red-300"
                />
              </div>
            ) : (
              <div className="flex gap-2">
                {state.gameId && (
                  <Link
                    to="/history/$id"
                    params={{ id: state.gameId }}
                    className="flex-1 rounded bg-neutral-100 px-3 py-2 text-center text-sm font-medium text-neutral-900 hover:bg-white"
                  >
                    Review game
                  </Link>
                )}
                <button
                  onClick={() => run(Setup)}
                  className="flex-1 rounded bg-neutral-800 px-3 py-2 text-sm hover:bg-neutral-700"
                >
                  New game
                </button>
              </div>
            )}
          </>
        )}
        {error && <p className="text-sm text-red-400">{error}</p>}
      </aside>
    </div>
  )
}

function SetupPanel(props: {
  mode: Mode
  setMode: (m: Mode) => void
  level: number
  setLevel: (l: number) => void
  color: ColorChoice
  setColor: (c: ColorChoice) => void
  hint?: string
  onStart: () => void
}) {
  const { mode, level, color } = props
  return (
    <div className="flex flex-col gap-4">
      <Field label="Mode">
        <Segmented
          value={mode}
          onChange={props.setMode}
          options={[
            ['bot', 'vs Stockfish'],
            ['local', '1v1'],
          ]}
        />
      </Field>

      {mode === 'bot' && (
        <>
          <Field label="Level">
            <div className="grid grid-cols-4 gap-1">
              {LEVELS.map((l) => (
                <button key={l} onClick={() => props.setLevel(l)} className={chip(level === l)}>
                  {l}
                </button>
              ))}
            </div>
          </Field>
          <Field label="Play as">
            <Segmented
              value={color}
              onChange={props.setColor}
              options={[
                ['white', 'White'],
                ['random', 'Random'],
                ['black', 'Black'],
              ]}
            />
          </Field>
        </>
      )}

      <div className="flex flex-col gap-1.5">
        {props.hint && <span className="text-xs text-neutral-500">{props.hint}</span>}
        <button
          onClick={props.onStart}
          className="rounded bg-neutral-100 px-3 py-2 text-sm font-medium text-neutral-900 hover:bg-white"
        >
          Start game
        </button>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-xs text-neutral-500">{label}</span>
      {children}
    </div>
  )
}

function Segmented<T extends string>({
  value,
  onChange,
  options,
}: {
  value: T
  onChange: (v: T) => void
  options: [T, string][]
}) {
  return (
    <div className="flex gap-1">
      {options.map(([v, label]) => (
        <button key={v} onClick={() => onChange(v)} className={`flex-1 ${chip(value === v)}`}>
          {label}
        </button>
      ))}
    </div>
  )
}

function chip(active: boolean) {
  return `rounded px-2 py-1.5 text-sm ${
    active ? 'bg-neutral-700 text-neutral-100' : 'bg-neutral-800/60 text-neutral-400 hover:text-neutral-100'
  }`
}

function status(s: match.State): string {
  const g = s.game
  const side = g.turn === 'white' ? 'White' : 'Black'
  switch (g.outcome) {
    case '*':
      if (s.botToMove) return 'Stockfish is thinking…'
      if (s.mode === 'bot') return g.check ? 'Your move — check' : 'Your move'
      return g.check ? `${side} to move — check` : `${side} to move`
    case '1-0':
      return `${g.method} — White wins`
    case '0-1':
      return `${g.method} — Black wins`
    default:
      return `Draw — ${g.method}`
  }
}
