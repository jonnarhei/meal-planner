type Props = {
    visual?: React.ReactNode
    title: string
    description: string
    actionLabel?: string
    onAction?: () => void
    busy?: boolean
}

function EmptyState({ visual, title, description, actionLabel, onAction, busy = false }: Props) {
    return (
        <div className="flex items-center justify-center py-16">
            <div className="flex flex-col items-center gap-3 text-center max-w-[340px]">
                {visual && <div className="mb-2">{visual}</div>}

                <span className="text-xl font-bold">{title}</span>
                <span className="text-sm text-stone-500 text-pretty">{description}</span>

                {actionLabel && onAction && (
                    <button
                        onClick={onAction}
                        disabled={busy}
                        className="mt-1.5 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-5 py-[11px] rounded-[10px] transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        {actionLabel}
                    </button>
                )}
            </div>
        </div>
    )
}

export default EmptyState
