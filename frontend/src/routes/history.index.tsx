import { useEffect, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { DeleteGame, ListGames } from '../../wailsjs/go/main/App'
import type { history } from '../../wailsjs/go/models'
import { ConfirmButton } from '../components/ConfirmButton'
import { errorText } from '../lib/errors'
import { formatDate, opponentLabel, resultLabel } from '../lib/games'

export const Route = createFileRoute('/history/')({
  component: History,
})

const TONE = {
  win: 'bg-green-500/15 text-green-400',
  loss: 'bg-red-500/15 text-red-400',
  draw: 'bg-neutral-700/50 text-neutral-300',
  neutral: 'bg-neutral-700/50 text-neutral-300',
}

function History() {
  const [games, setGames] = useState<history.Summary[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    ListGames().then(setGames).catch((e) => setError(errorText(e)))
  }, [])

  function remove(id: string) {
    DeleteGame(id)
      .then(() => setGames((gs) => gs?.filter((g) => g.id !== id) ?? null))
      .catch((e) => setError(errorText(e)))
  }

  if (error) return <p className="p-8 text-red-400">{error}</p>
  if (!games) return null

  return (
    <div className="mx-auto max-w-2xl p-8">
      <h1 className="mb-4 text-lg font-medium">History</h1>

      {games.length === 0 && (
        <p className="text-neutral-400">
          No games yet.{' '}
          <Link to="/play" className="text-neutral-100 underline underline-offset-4">
            Play one
          </Link>
        </p>
      )}

      <ul className="divide-y divide-neutral-800">
        {games.map((g) => {
          const r = resultLabel(g)
          return (
            <li key={g.id} className="group flex items-center">
              <Link
                to="/history/$id"
                params={{ id: g.id }}
                className="flex flex-1 items-center gap-4 rounded px-2 py-3 hover:bg-neutral-800/50"
              >
                <span className={`w-12 rounded px-2 py-0.5 text-center text-xs font-medium ${TONE[r.tone]}`}>
                  {r.text}
                </span>
                <span className="flex-1">
                  <span className="text-neutral-100">{opponentLabel(g)}</span>
                  <span className="block text-xs text-neutral-500">
                    {g.mode === 'bot' && `as ${g.humanColor} · `}
                    {Math.ceil(g.plies / 2)} moves · {g.method}
                  </span>
                </span>
                <span className="w-20 text-right text-sm text-neutral-400" title="Your accuracy">
                  {g.analyzed && g.mode === 'bot' ? `${Math.round(g.accuracy)}%` : ''}
                </span>
                <span className="w-28 text-right text-xs text-neutral-500">{formatDate(g.playedAt)}</span>
              </Link>
              <div className="ml-2 w-16 text-center">
                <ConfirmButton
                  label="Delete"
                  confirmLabel="Delete?"
                  onConfirm={() => remove(g.id)}
                  className="py-1 text-xs text-neutral-600 opacity-0 hover:text-neutral-300 focus:opacity-100 group-hover:opacity-100"
                  armedClassName="w-full rounded bg-red-500/20 py-1 text-xs text-red-300 hover:bg-red-500/30"
                />
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
