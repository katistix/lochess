// Display info for move grades produced by Go's analysis package.
// icon is shown in badges; glyph (bad moves only) is appended in the move list.
export const GRADES: Record<string, { label: string; plural: string; color: string; icon: string; glyph?: string }> = {
  best: { label: 'Best', plural: 'Best moves', color: '#81b64c', icon: '★' },
  excellent: { label: 'Excellent', plural: 'Excellent moves', color: '#96bc4b', icon: '!' },
  good: { label: 'Good', plural: 'Good moves', color: '#a3a3a3', icon: '✓' },
  inaccuracy: { label: 'Inaccuracy', plural: 'Inaccuracies', color: '#f7c631', icon: '?!', glyph: '?!' },
  mistake: { label: 'Mistake', plural: 'Mistakes', color: '#ffa459', icon: '?', glyph: '?' },
  blunder: { label: 'Blunder', plural: 'Blunders', color: '#fa412d', icon: '??', glyph: '??' },
}

export const GRADE_ORDER = ['best', 'excellent', 'good', 'inaccuracy', 'mistake', 'blunder']

// "1. e4" / "1... e5" for ply index i (0-based).
export function moveLabel(i: number, san: string) {
  const n = Math.floor(i / 2) + 1
  return i % 2 === 0 ? `${n}. ${san}` : `${n}... ${san}`
}
