import cburnettCss from '@lichess-org/chessground/assets/chessground.cburnett.css?raw'

export type BoardTheme = { id: string; name: string; light: string; dark: string }

// Flat two-colour boards (lichess's colours where it has an equivalent).
export const BOARDS: BoardTheme[] = [
  { id: 'brown', name: 'Brown', light: '#f0d9b5', dark: '#b58863' },
  { id: 'blue', name: 'Blue', light: '#dee3e6', dark: '#8ca2ad' },
  { id: 'green', name: 'Green', light: '#ffffdd', dark: '#86a666' },
  { id: 'olive', name: 'Olive', light: '#ececec', dark: '#c1c18e' },
  { id: 'purple', name: 'Purple', light: '#9f90b0', dark: '#7d4a8d' },
  { id: 'grey', name: 'Grey', light: '#dcdcdc', dark: '#a9a9a9' },
]

export type PieceSet = { id: string; name: string; author: string; license: string; note?: string }

export const PIECE_SETS: PieceSet[] = [
  { id: 'cburnett', name: 'Cburnett', author: 'Colin M.L. Burnett', license: 'GPLv2+' },
  { id: 'merida', name: 'Merida', author: 'Armando Hernandez Marroquin', license: 'GPLv2+' },
  { id: 'chessnut', name: 'Chessnut', author: 'Alexis Luengas', license: 'Apache 2.0' },
  { id: 'mpchess', name: 'MPChess', author: 'Maxime Chupin', license: 'GPLv3+' },
  { id: 'rhosgfx', name: 'RhosGFX', author: 'RhosGFX', license: 'CC0 1.0' },
  { id: 'kiwen-suwi', name: 'Kiwen-suwi', author: 'neverRare', license: 'CC BY 4.0' },
  { id: 'papercut', name: 'Papercut', author: 'Nikolay Anzarov', license: 'CC BY 4.0' },
  { id: 'fantasy', name: 'Fantasy', author: 'Maurizio Monge', license: 'MIT' },
  { id: 'spatial', name: 'Spatial', author: 'Maurizio Monge', license: 'MIT' },
  { id: 'celtic', name: 'Celtic', author: 'Maurizio Monge', license: 'MIT' },
  { id: 'shapes', name: 'Shapes', author: 'flugsio', license: 'CC BY-SA 4.0' },
  {
    id: 'mono',
    name: 'Mono',
    author: 'Thibault Duplessis and Colin M.L. Burnett',
    license: 'GPLv2+',
    note: 'Both sides look the same: a visualisation challenge.',
  },
]

export const DEFAULT_BOARD = 'brown'
export const DEFAULT_PIECES = 'cburnett'

const ROLES = { P: 'pawn', N: 'knight', B: 'bishop', R: 'rook', Q: 'queen', K: 'king' } as const
export type Role = keyof typeof ROLES

const files = import.meta.glob('../assets/pieces/*/*.svg', {
  eager: true,
  query: '?url',
  import: 'default',
}) as Record<string, string>

// cburnett ships inside chessground's CSS as data URIs; read them from there.
function cburnettUrl(role: Role, color: 'w' | 'b'): string {
  const re = new RegExp(`piece\\.${ROLES[role]}\\.${color === 'w' ? 'white' : 'black'}\\s*\\{[^}]*url\\('([^']+)'\\)`)
  return cburnettCss.match(re)?.[1] ?? ''
}

// URL of one piece image, e.g. pieceUrl('merida', 'N', 'w').
export function pieceUrl(set: string, role: Role, color: 'w' | 'b'): string {
  if (set === 'cburnett') return cburnettUrl(role, color)
  return files[`../assets/pieces/${set}/${color}${role}.svg`] ?? files[`../assets/pieces/${set}/${role}.svg`] ?? ''
}

export function boardTheme(id: string) {
  return BOARDS.find((b) => b.id === id) ?? BOARDS[0]
}

export function pieceSet(id: string) {
  return PIECE_SETS.find((p) => p.id === id) ?? PIECE_SETS[0]
}

// CSS that restyles every chessground board. The `html` prefix outranks
// chessground's own theme files, which may load after this style.
export function themeCss(boardId: string, setId: string): string {
  const b = boardTheme(boardId)
  const set = pieceSet(setId).id
  const rules = [
    `html cg-board { background-color: ${b.light}; background-image: repeating-conic-gradient(${b.dark} 0 25%, ${b.light} 0 50%); background-size: 25% 25%; }`,
    // Coordinates use the opposite square colour so they stay readable.
    `html .orientation-white .ranks :nth-child(odd), html .orientation-white .files :nth-child(even), html .orientation-black .ranks :nth-child(even), html .orientation-black .files :nth-child(odd) { color: ${b.dark}; }`,
    `html .orientation-white .ranks :nth-child(even), html .orientation-white .files :nth-child(odd), html .orientation-black .ranks :nth-child(odd), html .orientation-black .files :nth-child(even) { color: ${b.light}; }`,
  ]
  for (const role of Object.keys(ROLES) as Role[]) {
    for (const color of ['w', 'b'] as const) {
      const sel = `html .cg-wrap piece.${ROLES[role]}.${color === 'w' ? 'white' : 'black'}`
      // Double quotes: Vite's inlined SVG data URIs contain single quotes.
      rules.push(`${sel} { background-image: url("${pieceUrl(set, role, color)}"); }`)
    }
  }
  return rules.join('\n')
}
