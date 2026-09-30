import { useNavigate } from "react-router-dom"
import { useAppLayout } from "./AppLayout"

type Props = {
    title: string
    /** Right-hand slot. Defaults to the avatar that opens Profile. Pass null for nothing. */
    action?: React.ReactNode
}

/** Page title row for small screens, where the shell header is hidden. */
function MobileHeader({ title, action }: Props) {
    const navigate = useNavigate()
    const { user } = useAppLayout()

    const avatar = (
        <button
            onClick={() => navigate('/profile')}
            aria-label="Profile"
            className="w-11 h-11 -mr-1.5 flex items-center justify-center rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
        >
            <span className="w-8 h-8 rounded-full bg-stone-900 text-white flex items-center justify-center text-[13px] font-bold">
                {user?.email?.[0]?.toUpperCase() ?? '·'}
            </span>
        </button>
    )

    return (
        <div className="md:hidden flex items-center justify-between gap-3 mb-3 min-h-11">
            <h1 className="text-[26px] font-bold tracking-[-0.01em] leading-tight">{title}</h1>
            {action === undefined ? avatar : action}
        </div>
    )
}

export default MobileHeader
