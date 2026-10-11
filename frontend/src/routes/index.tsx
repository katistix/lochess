import { useEffect, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Ping } from '../../wailsjs/go/main/App'

export const Route = createFileRoute('/')({
  component: Home,
})

function Home() {
  const [backend, setBackend] = useState('…')

  useEffect(() => {
    Ping().then(setBackend).catch(() => setBackend('unreachable'))
  }, [])

  return (
    <div className="flex h-full flex-col items-center justify-center gap-2">
      <h1 className="text-3xl font-semibold">Hello, lochess</h1>
      <p className="text-sm text-neutral-400">backend: {backend}</p>
    </div>
  )
}
