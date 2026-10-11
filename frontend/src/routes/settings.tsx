import { createFileRoute } from '@tanstack/react-router'
import { Board } from '../components/Board'
import { useSettings } from '../lib/settings'
import { playSound } from '../lib/sound'
import { BOARDS, PIECE_SETS, boardTheme, pieceSet, pieceUrl } from '../lib/theme'

export const Route = createFileRoute('/settings')({
  component: Settings,
})

// A position with every piece type and a highlighted last move (Italian Game).
const PREVIEW_FEN = 'r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3'

function Settings() {
  const { settings, update } = useSettings()
  const current = pieceSet(settings.pieces)
  const board = boardTheme(settings.board)

  return (
    <div className="flex h-full items-center justify-center gap-10 p-6">
      <div className="aspect-square w-[min(75vh,48vw)]">
        <Board fen={PREVIEW_FEN} orientation="white" lastMove={['f1', 'c4']} />
      </div>

      <div className="flex w-80 flex-col gap-6">
        <section>
          <h2 className="mb-2 text-xs uppercase tracking-wide text-neutral-500">Board</h2>
          <div className="grid grid-cols-3 gap-2">
            {BOARDS.map((b) => (
              <Option key={b.id} label={b.name} selected={settings.board === b.id} onClick={() => update({ board: b.id })}>
                <div
                  className="aspect-[2/1] w-full rounded-sm"
                  style={{
                    backgroundImage: `repeating-conic-gradient(${b.dark} 0 25%, ${b.light} 0 50%)`,
                    backgroundSize: '25% 50%',
                  }}
                />
              </Option>
            ))}
          </div>
        </section>

        <section>
          <h2 className="mb-2 text-xs uppercase tracking-wide text-neutral-500">Pieces</h2>
          <div className="grid grid-cols-4 gap-2">
            {PIECE_SETS.map((p) => (
              <Option key={p.id} label={p.name} selected={settings.pieces === p.id} onClick={() => update({ pieces: p.id })}>
                {/* Shown on the chosen board's light square so dark pieces stay visible. */}
                <div className="flex h-10 items-center justify-center rounded-sm" style={{ backgroundColor: board.light }}>
                  <img src={pieceUrl(p.id, 'N', 'w')} alt="" className="h-9 w-9" />
                  <img src={pieceUrl(p.id, 'N', 'b')} alt="" className="-ml-2 h-9 w-9" />
                </div>
              </Option>
            ))}
          </div>
          <p className="mt-3 text-xs text-neutral-500">
            {current.note && <span className="block text-neutral-400">{current.note}</span>}
            {current.name} by {current.author} · {current.license}
          </p>
        </section>

        <section>
          <h2 className="mb-2 text-xs uppercase tracking-wide text-neutral-500">Sound</h2>
          <div className="flex items-center gap-3">
            <div className="flex gap-1">
              {([true, false] as const).map((on) => (
                <button
                  key={String(on)}
                  onClick={() => update({ sound: on })}
                  aria-pressed={settings.sound === on}
                  className={`rounded px-3 py-1.5 text-sm ${
                    settings.sound === on
                      ? 'bg-neutral-700 text-neutral-100'
                      : 'bg-neutral-800/60 text-neutral-400 hover:text-neutral-200'
                  }`}
                >
                  {on ? 'On' : 'Off'}
                </button>
              ))}
            </div>
            <input
              type="range"
              min={0}
              max={100}
              step={5}
              value={settings.volume}
              disabled={!settings.sound}
              onChange={(e) => update({ volume: Number(e.target.value) })}
              // Preview the new volume once the user lets go.
              onPointerUp={(e) => playSound('move', { sound: true, volume: Number(e.currentTarget.value) })}
              aria-label="Volume"
              className="flex-1 accent-neutral-300 disabled:opacity-40"
            />
            <button
              onClick={() => {
                playSound('move', settings)
                setTimeout(() => playSound('capture', settings), 450)
              }}
              disabled={!settings.sound}
              className="rounded bg-neutral-800/60 px-3 py-1.5 text-sm text-neutral-400 hover:text-neutral-200 disabled:opacity-40"
            >
              Test
            </button>
          </div>
        </section>

        <p className="text-xs text-neutral-600">Changes are saved automatically.</p>
      </div>
    </div>
  )
}

function Option({
  label,
  selected,
  onClick,
  children,
}: {
  label: string
  selected: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      aria-pressed={selected}
      title={label}
      className={`flex flex-col gap-1 rounded p-1.5 text-xs transition-colors ${
        selected
          ? 'bg-neutral-700 text-neutral-100 ring-2 ring-neutral-300'
          : 'bg-neutral-800/60 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-200'
      }`}
    >
      {children}
      <span className="truncate">{label}</span>
    </button>
  )
}
