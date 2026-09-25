import { useNavigate } from "react-router-dom"
import { useAuth } from "../context/AuthContext"
import { useEffect, useState } from "react"
import { updatePreferences } from "../api/user"
import toast from "react-hot-toast"
import { useAppLayout } from "../components/AppLayout"
import PreferencesForm, { emptyPreferences, type Preferences } from "../components/PreferencesForm"

const countLabel = (count: number, singular: string) =>
    `${count} ${singular}${count === 1 ? '' : 's'}`

const summarise = (preferences: Preferences) => {
    const { diets, intolerances, excluded } = preferences

    if (diets.length === 0 && intolerances.length === 0 && excluded.length === 0) {
        return 'No food preferences set'
    }

    return [
        countLabel(diets.length, 'diet'),
        countLabel(intolerances.length, 'intolerance'),
        countLabel(excluded.length, 'excluded ingredient'),
    ].join(', ')
}

function Profile() {
    const { setToken } = useAuth()
    const navigate = useNavigate()
    const { user, userLoading, setUser } = useAppLayout()

    const [preferences, setPreferences] = useState<Preferences>(emptyPreferences)
    const [saving, setSaving] = useState(false)

    useEffect(() => {
        if (!user) return

        setPreferences({
            diets: user.dietary_preferences ?? [],
            intolerances: user.intolerances ?? [],
            excluded: user.excluded_ingredients ?? [],
        })
    }, [user])

    const handleSave = async () => {
        setSaving(true)
        try {
            await updatePreferences(preferences.diets, preferences.intolerances, preferences.excluded)
            setUser(prev => prev && ({
                ...prev,
                dietary_preferences: preferences.diets,
                intolerances: preferences.intolerances,
                excluded_ingredients: preferences.excluded,
            }))
            toast.success('Preferences saved!')
        } catch (err) {
            toast.error('Failed to save preferences')
        } finally {
            setSaving(false)
        }
    }

    const handleLogout = () => {
        setToken(null)
        navigate('/login')
    }

    return (
        <div className="grid grid-cols-[300px_minmax(0,640px)] gap-8 items-start">

            {/* Account */}
            <div className="border border-stone-200 rounded-[20px] p-6 flex flex-col gap-[18px]">
                <div className="flex items-center gap-3.5">
                    <span className="w-14 h-14 flex-none rounded-full bg-orange-100 text-orange-700 flex items-center justify-center text-[22px] font-bold">
                        {user?.email?.[0]?.toUpperCase() ?? '·'}
                    </span>
                    <div className="flex flex-col gap-0.5 min-w-0">
                        <span className="text-[13px] text-stone-500">Signed in as</span>
                        <span className="text-[15px] font-semibold break-words">
                            {userLoading ? '…' : user?.email ?? 'Unknown'}
                        </span>
                    </div>
                </div>

                <div className="h-px bg-stone-100" />

                <span className="text-sm text-stone-600">{summarise(preferences)}</span>

                <button
                    onClick={handleLogout}
                    className="w-full bg-red-50 hover:bg-red-100 text-red-600 text-sm font-semibold py-2.5 rounded-[10px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                >
                    Sign out
                </button>
            </div>

            {/* Food preferences */}
            <div className="flex flex-col gap-7">
                <div className="flex flex-col gap-1">
                    <span className="text-xl font-bold">Food preferences</span>
                    <span className="text-sm text-stone-500">
                        Used when a new meal plan is generated or you press Surprise me.
                    </span>
                </div>

                <PreferencesForm value={preferences} onChange={setPreferences} />

                <div className="border-t border-stone-100 pt-6">
                    <button
                        onClick={handleSave}
                        disabled={saving || userLoading}
                        className="bg-orange-500 hover:bg-orange-600 text-white text-[15px] font-semibold px-[22px] py-[11px] rounded-[10px] transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        {saving ? 'Saving…' : 'Save preferences'}
                    </button>
                </div>
            </div>
        </div>
    )
}


export default Profile
