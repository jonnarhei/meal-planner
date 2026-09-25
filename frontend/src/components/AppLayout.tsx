import { useEffect, useState } from "react"
import { NavLink, Outlet, useLocation, useNavigate, useOutletContext } from "react-router-dom"
import toast from "react-hot-toast"
import { useAuth } from "../context/AuthContext"
import { getCurrentUser } from "../api/user"
import type { User } from "../api/types"

export type AppLayoutContext = {
    user: User | null
    userLoading: boolean
    setUser: React.Dispatch<React.SetStateAction<User | null>>
    setWeekLabel: (label: string | null) => void
}

export function useAppLayout() {
    return useOutletContext<AppLayoutContext>()
}

const navItems = [
    { to: '/meal-plan', label: 'Meal plan' },
    { to: '/recipes', label: 'Recipes' },
    { to: '/shopping-list', label: 'Shopping list' },
]

const pageTitle = (pathname: string) => {
    if (pathname.startsWith('/meal-plan')) return 'Meal plan'
    if (pathname.startsWith('/recipes')) return 'Recipes'
    if (pathname.startsWith('/shopping-list')) return 'Shopping list'
    if (pathname.startsWith('/profile')) return 'Profile'
    return 'Meal Planner'
}

function AppLayout() {
    const { setToken } = useAuth()
    const navigate = useNavigate()
    const { pathname } = useLocation()

    const [user, setUser] = useState<User | null>(null)
    const [userLoading, setUserLoading] = useState(true)
    const [weekLabel, setWeekLabel] = useState<string | null>(null)

    useEffect(() => {
        const fetchUser = async () => {
            try {
                setUser(await getCurrentUser())
            } catch (err) {
                toast.error('Error fetching user data')
            } finally {
                setUserLoading(false)
            }
        }

        fetchUser()
    }, [])

    const handleLogout = () => {
        setToken(null)
        navigate('/login')
    }

    const onProfile = pathname.startsWith('/profile')
    const initial = user?.email?.[0]?.toUpperCase() ?? '·'

    return (
        <div className="h-screen flex flex-col bg-white text-stone-900">

            <header className="h-[68px] flex-none flex items-center justify-between px-7 bg-white border-b border-stone-200">
                <div className="flex items-center gap-3.5">
                    <span className="w-2.5 h-2.5 rounded-[3px] bg-orange-500" />
                    <h1 className="text-[22px] font-bold tracking-[-0.01em]">{pageTitle(pathname)}</h1>
                </div>

                {weekLabel && (
                    <div className="flex items-center gap-1 bg-stone-100 rounded-[10px] p-1">
                        <button
                            disabled
                            aria-label="Previous week"
                            className="w-[30px] h-[30px] rounded-[7px] text-base text-stone-600 disabled:cursor-default"
                        >
                            ‹
                        </button>
                        <span className="text-sm font-semibold px-2.5">{weekLabel}</span>
                        <button
                            disabled
                            aria-label="Next week"
                            className="w-[30px] h-[30px] rounded-[7px] text-base text-stone-600 disabled:cursor-default"
                        >
                            ›
                        </button>
                    </div>
                )}

                <button
                    onClick={() => navigate('/profile')}
                    className={`flex items-center gap-2 text-sm rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300
                        ${onProfile
                            ? 'bg-stone-100 text-stone-900 font-semibold py-1 pl-1 pr-3.5'
                            : 'text-stone-700 py-1 pl-1 pr-3.5 hover:bg-stone-100'
                        }`}
                >
                    <span className="w-8 h-8 rounded-full bg-stone-900 text-white flex items-center justify-center text-[13px] font-bold">
                        {initial}
                    </span>
                    Profile
                </button>
            </header>

            <div className="flex-1 flex min-h-0">
                <aside className="w-[210px] flex-none px-3 py-5 flex flex-col gap-1 bg-orange-50">
                    {navItems.map(item => (
                        <NavLink
                            key={item.to}
                            to={item.to}
                            className={({ isActive }) =>
                                `px-3.5 py-2.5 rounded-[10px] text-[15px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300
                                ${isActive
                                    ? 'bg-white text-stone-900 font-bold shadow-[0_1px_3px_rgba(120,53,15,0.12)]'
                                    : 'text-stone-600 font-medium hover:text-stone-900'
                                }`
                            }
                        >
                            {item.label}
                        </NavLink>
                    ))}

                    <button
                        onClick={handleLogout}
                        className="mt-auto text-left px-3.5 py-2.5 rounded-[10px] text-sm text-red-600 transition-colors hover:bg-red-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        Sign out
                    </button>
                </aside>

                <main className="flex-1 overflow-auto bg-white py-7 px-8">
                    <Outlet context={{ user, userLoading, setUser, setWeekLabel } satisfies AppLayoutContext} />
                </main>
            </div>
        </div>
    )
}

export default AppLayout
