import { useNavigate } from "react-router-dom"
import { useAuth } from "../context/AuthContext"
import HamburgerMenu from "./HamburgerMenu"
import { useEffect, useState } from "react"
import { getCurrentUser, updatePreferences } from "../api/user"
import { DIETARY_OPTIONS, INTOLERANCE_OPTIONS, type User as UserType } from "../api/types"
import toast from "react-hot-toast"

function Profile() {
    const { setToken } = useAuth()
    const navigate = useNavigate()
    const [user, setUser] = useState<UserType | null>(null)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [selectedDietaryPreferences, setSelectedDietaryPreferences] = useState<string[]>([])
    const [selectedIntolerances, setSelectedIntolerances] = useState<string[]>([])
    const [excludedIngredients, setExcludedIngredients] = useState<string[]>([])
    const [ingredientInput, setIngredientInput] = useState('')



    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const data = await getCurrentUser()
                setUser(data)
                setSelectedDietaryPreferences(data.dietary_preferences ?? [])
                setSelectedIntolerances(data.intolerances ?? [])
                setExcludedIngredients(data.excluded_ingredients)
            } catch (err) {
                toast.error("Error fetching user data")
            } finally {
                setLoading(false)
            }
        }

        fetchUserData()
    }, [])

    const toggleDietaryPreference = (option: string) => {
        setSelectedDietaryPreferences(prev => 
            prev.includes(option) ? prev.filter(p => p !== option) : [...prev, option]
        )
    }

    const toggleIntolerance = (option: string) => {
        setSelectedIntolerances(prev => 
            prev.includes(option) ? prev.filter(p => p !== option) : [...prev, option]
        )
    }

    const addIngredient = () => {
        const trimmed = ingredientInput.trim().toLowerCase()
        if (!trimmed) return
        if (excludedIngredients.includes(trimmed)) {
            toast.error('Already added')
            return
        }
        setExcludedIngredients(prev => [...prev, trimmed])
        setIngredientInput('')
    }

    const removeIngredient = (ingredient: string) => {
        setExcludedIngredients(prev => prev.filter(i => i !== ingredient.trim()))
    }

    const handleSave = async () => {
        setSaving(true)
        try {
            await updatePreferences(selectedDietaryPreferences, selectedIntolerances, excludedIngredients)
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
                                onClick={() => toggleDietaryPreference(option)}
                                className={`px-4 py-2 rounded-xl text-sm font-medium border transition-colors
                                    ${selectedDietaryPreferences.includes(option)
                                        ? 'bg-orange-500 text-white border-orange-500'
                                        : 'bg-white text-gray-600 border-gray-200 hover:border-orange-300'
                                    }`}
                            >
                                {option}
                            </button>
                        ))}
                    </div>

                    {/* Intolerances */}
                    <h2 className="text-sm font-semibold text-gray-700 mb-3">Intolerances</h2>
                    <div className="flex flex-wrap gap-3 mb-6">
                        {INTOLERANCE_OPTIONS.map(option => (
                            <button
                                key={option}
                                onClick={() => toggleIntolerance(option)}
                                className={`px-4 py-2 rounded-xl text-sm font-medium border transition-colors
                                    ${selectedIntolerances.includes(option)
                                        ? 'bg-red-400 text-white border-red-400'
                                        : 'bg-white text-gray-600 border-gray-200 hover:border-red-300'
                                    }`}
                            >
                                {option}
                            </button>
                        ))}
                    </div>

                    {/* Excluded ingredients */}
                    <h2 className="text-sm font-semibold text-gray-700 mb-1">Exclude Ingredients</h2>
                    <p className="text-xs text-gray-400 mb-3">Add any allergies we have missed or other ingredients you don't want</p>
                    <div className="flex gap-2 mb-3">
                        <input
                            type="text"
                            placeholder="e.g. mushrooms"
                            value={ingredientInput}
                            onChange={e => setIngredientInput(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && addIngredient()}
                            className="flex-1 border border-gray-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                        />
                        <button
                            onClick={addIngredient}
                            className="bg-orange-100 hover:bg-orange-200 text-orange-600 font-medium px-4 py-2 rounded-xl text-sm transition-colors"
                        >
                            Add
                        </button>
                    </div>
                    {excludedIngredients.length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-6">
                            {excludedIngredients.map(ingredient => (
                                <span
                                    key={ingredient}
                                    className="flex items-center gap-1 bg-gray-100 text-gray-600 text-sm px-3 py-1 rounded-xl"
                                >
                                    {ingredient}
                                    <button
                                        onClick={() => removeIngredient(ingredient)}
                                        className="text-gray-400 hover:text-red-400 transition-colors ml-1"
                                    >
                                        ×
                                    </button>
                                </span>
                            ))}
                        </div>
                    )}

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