import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { updatePreferences } from "../api/user";
import { DIETARY_OPTIONS } from "../api/types";
import toast from "react-hot-toast";

function DietaryPreferences() {
    const [selected, setSelected] = useState<string[]>([])
    const [loading, setLoading] = useState(false)

    const navigate = useNavigate()

    const togglePreference = (option: string) => {
        setSelected(prev => 
            prev.includes(option) ? prev.filter(p => p !== option) : [...prev, option]
        )
    }

    const handleSubmit = async () => {
        setLoading(true)
        try {
            updatePreferences(selected)
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
                <p className="text-gray-500 text-sm mb-6">Select any that apply — we'll use these to tailor your meal plan.</p>

                <div className="flex flex-wrap gap-3 mb-8">
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