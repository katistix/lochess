import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/play')({
  component: Play,
})

function Play() {
  return (
    <div className="flex h-full items-center justify-center text-neutral-400">
      Board goes here.
    </div>
  )
}
