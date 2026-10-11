// Turns an error from a Go call into a sentence for the user.
export function errorText(e: unknown): string {
  const msg = String(e).replace(/^Error:\s*/, '')
  if (msg.includes('game not found')) return 'This game no longer exists.'
  if (msg.includes('fairy-stockfish not found')) {
    return 'Stockfish is not installed. Install it with: brew install fairy-stockfish'
  }
  if (msg.startsWith('bot:') || msg.includes('engine exited')) {
    return 'The engine stopped responding. Try again.'
  }
  return msg.charAt(0).toUpperCase() + msg.slice(1)
}
