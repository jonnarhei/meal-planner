import { useEffect, useRef, useState } from "react"

type Props = {
    busy: boolean
    onSurprise: () => void
    onUseOwn: () => void
}

function ChangeRecipePopover({ busy, onSurprise, onUseOwn }: Props) {
    const [open, setOpen] = useState(false)
    const popoverRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        if (!open) return

        const handleClickOutside = (e: MouseEvent) => {
            if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
                setOpen(false)
            }
        }
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') setOpen(false)
        }

        document.addEventListener('mousedown', handleClickOutside)
        document.addEventListener('keydown', handleEscape)

        return () => {
            document.removeEventListener('mousedown', handleClickOutside)
            document.removeEventListener('keydown', handleEscape)
        }
    }, [open])

    const choose = (action: () => void) => {
        setOpen(false)
        action()
    }

    return (
        <div className="relative" ref={popoverRef}>
            <button
                onClick={() => setOpen(prev => !prev)}
                disabled={busy}
                className="text-xs bg-orange-100 hover:bg-orange-200 text-orange-600 font-medium px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50"
            >
                {busy ? 'Changing...' : 'Change'}
            </button>

            {open && (
                <div className="absolute bottom-full right-0 mb-3 w-44 bg-white rounded-xl shadow-lg border border-orange-100 p-2 flex flex-col gap-1.5 z-20">
                    <button
                        onClick={() => choose(onSurprise)}
                        className="w-full bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-3 py-2 rounded-lg transition-colors"
                    >
                        Surprise me
                    </button>
                    <button
                        onClick={() => choose(onUseOwn)}
                        className="w-full bg-orange-100 hover:bg-orange-200 text-orange-600 text-sm font-semibold px-3 py-2 rounded-lg transition-colors"
                    >
                        Use my own
                    </button>

                    {/* tail */}
                    <span className="absolute -bottom-1.5 right-5 w-3 h-3 bg-white border-r border-b border-orange-100 rotate-45"></span>
                </div>
            )}
        </div>
    )
}

export default ChangeRecipePopover