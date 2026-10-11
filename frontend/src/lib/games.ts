import type { history } from '../../wailsjs/go/models'

type GameLike = Pick<history.Summary, 'mode' | 'level' | 'humanColor' | 'result'>

export function opponentLabel(g: GameLike) {
  return g.mode === 'bot' ? `Stockfish level ${g.level}` : 'Local 1v1'
}

// Result from the human's side; local games just show the score.
export function resultLabel(g: GameLike): { text: string; tone: 'win' | 'loss' | 'draw' | 'neutral' } {
  if (g.result === '1/2-1/2') return { text: 'Draw', tone: 'draw' }
  if (g.mode !== 'bot') return { text: g.result, tone: 'neutral' }
  const won = (g.result === '1-0') === (g.humanColor === 'white')
  return won ? { text: 'Win', tone: 'win' } : { text: 'Loss', tone: 'loss' }
}

export function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
