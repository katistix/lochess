import { useEffect, useState } from 'react'

type Props = {
  label: string
  confirmLabel: string
  onConfirm: () => void
  className?: string // idle style
  armedClassName?: string // style while waiting for the second click
  timeout?: number // ms before an armed button resets
}

// Two-click button for destructive actions: the first click arms it, the
// second runs it. It disarms on its own after `timeout` or when focus leaves.
export function ConfirmButton({ label, confirmLabel, onConfirm, className, armedClassName, timeout = 3000 }: Props) {
  const [armed, setArmed] = useState(false)

  useEffect(() => {
    if (!armed) return
    const t = setTimeout(() => setArmed(false), timeout)
    return () => clearTimeout(t)
  }, [armed, timeout])

  return (
    <button
      onClick={() => {
        if (!armed) return setArmed(true)
        setArmed(false)
        onConfirm()
      }}
      onBlur={() => setArmed(false)}
      className={armed ? armedClassName : className}
    >
      {armed ? confirmLabel : label}
    </button>
  )
}
