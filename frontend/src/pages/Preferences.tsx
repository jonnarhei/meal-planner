import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { updatePreferences } from "../api/user";
import { DIETARY_OPTIONS, INTOLERANCE_OPTIONS } from "../api/types";
import toast from "react-hot-toast";

function DietaryPreferences() {
    const [selectedDietaryPreferences, setSelectedDietaryPreferences] = useState<string[]>([])
    const [selectedIntolerances, setSelectedIntolerances] = useState<string[]>([])
    const [excludedIngredients, setExcludedIngredients] = useState<string[]>([])
    const [ingredientInput, setIngredientInput] = useState('')
    const [loading, setLoading] = useState(false)


    const navigate = useNavigate()

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
        setExcludedIngredients(prev => prev.filter(i => i !== ingredient))
    }

    const handleSubmit = async () => {
        setLoading(true)
        try {
            updatePreferences(selectedDietaryPreferences, selectedIntolerances, excludedIngredients)
            navigate('/meal-plan')
        } catch (err) {
            toast.error('Failed to save preferences')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <div className="bg-white rounded-3xl shadow-md border border-orange-100 p-8 w-full max-w-md">
                <h1 className="text-2xl font-bold text-orange-600 mb-2">Dietary Preferences</h1>
                <p className="text-gray-500 text-sm mb-6">Help us tailor your meal plan to your needs.</p>

                {/* Diet */}
                <h2 className="text-sm font-semibold text-gray-700 mb-3">Diet</h2>
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
                <h2 className="text-sm font-semibold text-gray-700 mb-3">Exclude Ingredients</h2>
                <p className="text-xs text-gray-400 mb-3">Add any allergies we have missed or other ingredients you don't want</p>
                <div className="flex gap-2 mb-6">
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
                    onClick={handleSubmit}
                    disabled={loading}
                    className="w-full bg-orange-500 hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition-colors disabled:opacity-50"
                >
                    {loading ? 'Saving...' : 'Continue'}
                </button>

                <button
                    onClick={() => navigate('/meal-plan')}
                    className="w-full text-gray-400 hover:text-gray-600 text-sm mt-3 py-2 transition-colors"
                >
                    Skip for now
                </button>
            </div>
        </div>
    )
}

export default DietaryPreferences