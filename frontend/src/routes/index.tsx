import { createFileRoute, redirect } from '@tanstack/react-router'

// The app opens straight onto the board.
export const Route = createFileRoute('/')({
  beforeLoad: () => {
    throw redirect({ to: '/play' })
  },
})
