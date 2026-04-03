import { useNavigate } from "react-router-dom"
import { useAuth } from "../context/AuthContext"
import HamburgerMenu from "./HamburgerMenu"
import { useEffect, useState } from "react"
import { getCurrentUser, updatePreferences } from "../api/user"
import { DIETARY_OPTIONS, type User as UserType } from "../api/types"
import toast from "react-hot-toast"

function Profile() {
    const { setToken } = useAuth()
    const navigate = useNavigate()
    const [user, setUser] = useState<UserType | null>(null)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [selected, setSelected] = useState<string[]>([])



    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const data = await getCurrentUser()
                console.log('user data:', data)
                console.log('preferences:', data.dietary_preferences)
                setUser(data)
                setSelected(data.dietary_preferences ?? [])
            } catch (err) {
                toast.error("Error fetching user data")
            } finally {
                setLoading(false)
            }
        }

        fetchUserData()
    }, [])

    const togglePreference = (option: string) => {
        setSelected(prev => 
            prev.includes(option) ? prev.filter(p => p !== option) : [...prev, option]
        )
    }

    const handleSave = async () => {
        setSaving(true)
        try {
            await updatePreferences(selected)
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

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500">Loading...</p>
        </div>
    )

    return (
        <div className="min-h-screen bg-orange-50">
            <div className="bg-white shadow-sm">
                <div className="max-w-screen-2xl mx-auto px-6 py-4 flex items-center gap-4">
                    <HamburgerMenu />
                    <h1 className="text-2xl font-bold text-orange-600">Profile</h1>
                </div>
            </div>

            <div className="max-w-lg mx-auto px-6 py-12">
                <div className="bg-white rounded-3xl shadow-md border border-orange-100 p-8">
                    
                    {/* User info */}
                    <div className="flex items-center gap-4 mb-8">
                        <div className="w-16 h-16 rounded-full bg-orange-100 flex items-center justify-center">
                            <span className="text-orange-400 font-bold text-xl">
                                {user?.email?.[0].toUpperCase()}
                            </span>
                        </div>
                        <div>
                            <p className="text-gray-800 font-semibold">{user?.email}</p>
                        </div>
                    </div>

                    {/* Dietary preferences */}
                    <h2 className="text-sm font-semibold text-gray-700 mb-3">Dietary Preferences</h2>

                    <div className="flex flex-wrap gap-3 mb-6">
                        {DIETARY_OPTIONS.map(option => (
                            <button
                                key={option}
                                onClick={() => togglePreference(option)}
                                className={`px-4 py-2 rounded-xl text-sm font-medium border transition-colors
                                    ${selected.includes(option)
                                        ? 'bg-orange-500 text-white border-orange-500'
                                        : 'bg-white text-gray-600 border-gray-200 hover:border-orange-300'
                                    }`}
                            >
                                {option}
                            </button>
                        ))}
                    </div>

                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="w-full bg-orange-500 hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition-colors disabled:opacity-50 mb-4"
                    >
                        {saving ? 'Saving...' : 'Save Preferences'}
                    </button>

                    <hr className="border-orange-100 my-4" />

                    <button
                        onClick={handleLogout}
                        className="w-full bg-red-50 hover:bg-red-100 text-red-500 font-semibold py-3 rounded-xl transition-colors"
                    >
                        Sign out
                    </button>
                </div>
            </div>
        </div>
    )
}


export default Profile