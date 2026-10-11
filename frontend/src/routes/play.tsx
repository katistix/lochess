import { useEffect, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { GameState, Move, NewGame } from '../../wailsjs/go/main/App'
import type { game } from '../../wailsjs/go/models'
import { Board } from '../components/Board'

export const Route = createFileRoute('/play')({
  component: Play,
})

function Play() {
  const [state, setState] = useState<game.State | null>(null)

  useEffect(() => {
    GameState().then(setState)
  }, [])

  function handleMove(from: string, to: string) {
    // On rejection, re-sync so the board snaps back to Go's position.
    Move(from + to)
      .then(setState)
      .catch(() => GameState().then(setState))
  }

  if (!state) return null

  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 p-4">
      <div className="aspect-square w-[min(75vh,90vw)]">
        <Board state={state} onMove={handleMove} />
      </div>
      <div className="flex items-center gap-4">
        <p className="text-sm text-neutral-300">{status(state)}</p>
        <button
          onClick={() => NewGame().then(setState)}
          className="rounded bg-neutral-800 px-3 py-1.5 text-sm hover:bg-neutral-700"
        >
          New game
        </button>
      </div>
    </div>
  )
}

function status(s: game.State): string {
  const side = s.turn === 'white' ? 'White' : 'Black'
  switch (s.outcome) {
    case '*':
      return s.check ? `${side} to move — check` : `${side} to move`
    case '1-0':
      return `${s.method} — White wins`
    case '0-1':
      return `${s.method} — Black wins`
    default:
      return `Draw — ${s.method}`
  }
}
