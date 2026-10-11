import { Link, Outlet, createRootRoute } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: RootLayout,
})

const navLink = 'rounded px-3 py-1.5 text-sm text-neutral-400 hover:text-neutral-100'
const navLinkActive = 'bg-neutral-800 text-neutral-100'

function RootLayout() {
  return (
    <div className="flex h-screen flex-col bg-neutral-900 text-neutral-100">
      <nav className="flex gap-1 border-b border-neutral-800 px-4 py-2">
        <Link to="/" className={navLink} activeProps={{ className: navLinkActive }} activeOptions={{ exact: true }}>
          Home
        </Link>
        <Link to="/play" className={navLink} activeProps={{ className: navLinkActive }}>
          Play
        </Link>
        <Link to="/history" className={navLink} activeProps={{ className: navLinkActive }}>
          History
        </Link>
      </nav>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
