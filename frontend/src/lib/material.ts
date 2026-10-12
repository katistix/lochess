import type { Color } from '@lichess-org/chessground/types'

export type Role = 'pawn' | 'bishop' | 'knight' | 'rook' | 'queen'
export type Captured = [Role, number][]

// Shown in this order, as on chess.com.
const ROLES: [Role, string, number, number][] = [
  // role, FEN letter, value, starting count
  ['pawn', 'p', 1, 8],
  ['bishop', 'b', 3, 2],
  ['knight', 'n', 3, 2],
  ['rook', 'r', 5, 2],
  ['queen', 'q', 9, 1],
]

// What each side has taken and how far ahead it is, read from the board alone.
// A promoted piece can hide a capture of the same kind; the score stays exact.
export function material(fen: string): Record<Color, { captured: Captured; advantage: number }> {
  const board = fen.split(' ')[0]
  const count = (ch: string) => board.split(ch).length - 1
  let diff = 0 // White's material minus Black's
  const white: Captured = [] // black pieces White has taken
  const black: Captured = []
  for (const [role, letter, value, start] of ROLES) {
    const w = count(letter.toUpperCase())
    const b = count(letter)
    diff += (w - b) * value
    if (b < start) white.push([role, start - b])
    if (w < start) black.push([role, start - w])
  }
  return {
    white: { captured: white, advantage: Math.max(0, diff) },
    black: { captured: black, advantage: Math.max(0, -diff) },
  }
}
