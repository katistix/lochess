import { useEffect, useRef, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import type { Color } from '@lichess-org/chessground/types'
import { AgreeDraw, BotMove, GameState, Move, NewGame, Resign, Setup, Takeback } from '../../wailsjs/go/main/App'
import type { match } from '../../wailsjs/go/models'
import { Board } from '../components/Board'
import { ConfirmButton } from '../components/ConfirmButton'
import { MoveList } from '../components/MoveList'
import { PlayerBar } from '../components/PlayerBar'
import { PromotionPicker } from '../components/PromotionPicker'
import { errorText } from '../lib/errors'
import { useSettings } from '../lib/settings'
import { moveSound, playSound } from '../lib/sound'

export const Route = createFileRoute('/play')({
  component: Play,
})

type Mode = 'bot' | 'local'
type ColorChoice = 'white' | 'random' | 'black'

const LEVELS = [1, 2, 3, 4, 5, 6, 7, 8]

function Play() {
  const [state, setState] = useState<match.State | null>(null)
  const [error, setError] = useState('')
  // The setup choices are saved with the settings so they survive restarts.
  const { settings, update } = useSettings()
  const mode = settings.mode as Mode
  const level = settings.level
  const color = settings.color as ColorChoice
  // A pawn move to the last rank waiting for the user to pick a piece.
  const [promoting, setPromoting] = useState<{ from: string; to: string } | null>(null)

  // Bumped on every new game so replies from a previous game are dropped.
  const gen = useRef(0)
  function accept(g: number) {
    return (s: match.State) => {
      if (g === gen.current) setState(s)
    }
  }
  // On error, show it and re-sync so the board snaps back to Go's position.
  function resync(g: number) {
    return (err: unknown) => {
      setError(errorText(err))
      GameState().then(accept(g))
    }
  }

  useEffect(() => {
    GameState().then(setState)
  }, [])

  // Sounds follow Go's state: a new move, a game starting or ending.
  // The first state after opening the page is silent.
  const prev = useRef<match.State | null>(null)
  useEffect(() => {
    const before = prev.current
    prev.current = state
    if (!before || !state) return
    const moves = state.game.moves
    const started = state.phase === 'playing' && before.phase !== 'playing'
    const ended = state.phase === 'over' && before.phase !== 'over'
    // A game starts before its first move and ends after its last.
    if (started) playSound('notify', settings)
    if (moves.length === before.game.moves.length + 1) playSound(moveSound(moves[moves.length - 1]), settings)
    if (ended) playSound('notify', settings)
  }, [state])

  // Whenever it's the bot's turn, ask Go to play. Duplicate calls are no-ops in Go.
  // A failure only shows the error (with a retry button): re-syncing here
  // would set a new state and call the bot again in a loop.
  useEffect(() => {
    if (!state?.botToMove) return
    const g = gen.current
    BotMove()
      .then(accept(g))
      .catch((err) => g === gen.current && setError(errorText(err)))
  }, [state])

  function startGame() {
    const g = ++gen.current
    setError('')
    NewGame({ mode, level, color }).then(accept(g)).catch(resync(g))
  }

  function handleMove(from: string, to: string) {
    setError('')
    if (state?.phase === 'playing' && state.game.promotions.includes(from + to)) {
      setPromoting({ from, to })
      return
    }
    play(from + to)
  }

  function play(uci: string) {
    // In setup, the first move as White starts the game with the chosen options.
    const g = state?.phase === 'setup' ? ++gen.current : gen.current
    const started = state?.phase === 'setup' ? NewGame({ mode, level, color }) : Promise.resolve()
    started
      .then(() => Move(uci))
      .then(accept(g))
      .catch(resync(g))
  }

  function pickPromotion(piece: string) {
    if (!promoting) return
    setPromoting(null)
    play(promoting.from + promoting.to + piece)
  }

  function cancelPromotion() {
    setPromoting(null)
    GameState().then(accept(gen.current)) // fresh state puts the pawn back
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
  if (state.phase === 'playing' && !state.botToMove && !promoting) movableColor = g.turn as Color

  // Player strips above and below the board; the dot marks whose move it is.
  const top: Color = orientation === 'white' ? 'black' : 'white'
  const vsBot = setup ? mode === 'bot' : state.mode === 'bot'
  const botLabel = `Stockfish level ${setup ? level : state.level}`
  const playing = state.phase === 'playing'
  function bar(c: Color) {
    let label = c === 'white' ? 'White' : 'Black'
    if (vsBot) label = c === orientation ? 'You' : botLabel
    return <PlayerBar fen={g.fen} color={c} label={label} active={playing && g.turn === c} />
  }

  return (
    <div className="flex h-full items-center justify-center gap-8 p-6">
      <div className="flex flex-col gap-1">
        {bar(top)}
        <div className="relative aspect-square w-[min(72vh,65vw)]">
          {promoting && (
            <PromotionPicker
              square={promoting.to}
              color={g.turn as Color}
              orientation={orientation}
              onPick={pickPromotion}
              onCancel={cancelPromotion}
            />
          )}
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
        {bar(orientation)}
      </div>

      <aside className="flex w-64 flex-col gap-5">
        {setup ? (
          <SetupPanel
            mode={mode}
            setMode={(m) => update({ mode: m })}
            level={level}
            setLevel={(l) => update({ level: l })}
            color={color}
            setColor={(c) => update({ color: c })}
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
              <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
                {state.botToMove && error ? (
                  <button onClick={() => run(BotMove)} className="text-sm text-neutral-200 hover:text-white">
                    Retry bot move
                  </button>
                ) : (
                  <button
                    onClick={() => run(Takeback)}
                    disabled={state.botToMove || g.moves.length < (state.mode === 'bot' ? 2 : 1)}
                    className="text-sm text-neutral-500 hover:text-neutral-200 disabled:pointer-events-none disabled:opacity-40"
                  >
                    Takeback
                  </button>
                )}
                {state.mode === 'local' && (
                  <ConfirmButton
                    label="Draw"
                    confirmLabel="Click again to agree a draw"
                    onConfirm={() => run(AgreeDraw)}
                    className="text-sm text-neutral-500 hover:text-neutral-200"
                    armedClassName="rounded bg-neutral-700 px-2 py-1 text-sm text-neutral-100"
                  />
                )}
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
