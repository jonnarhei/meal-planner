import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { updatePreferences } from "../api/user";
import toast from "react-hot-toast";
import PreferencesForm, { emptyPreferences, type Preferences } from "../components/PreferencesForm";

function DietaryPreferences() {
    const [preferences, setPreferences] = useState<Preferences>(emptyPreferences)
    const [loading, setLoading] = useState(false)

    const navigate = useNavigate()

    const handleSubmit = async () => {
        setLoading(true)
        try {
            await updatePreferences(preferences.diets, preferences.intolerances, preferences.excluded)
            navigate('/meal-plan')
        } catch (err) {
            toast.error('Failed to save preferences')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-orange-50 flex flex-col text-stone-900">

            <div className="flex items-center gap-2.5 px-8 py-6">
                <span className="w-2.5 h-2.5 rounded-[3px] bg-orange-500" />
                <span className="text-base font-bold">Meal Planner</span>
            </div>

            <div className="flex-1 flex justify-center px-8 pb-8">
                <div className="w-full max-w-[520px] self-start bg-white border border-orange-200 rounded-3xl p-8 flex flex-col gap-6">
                    <div className="flex flex-col gap-1.5">
                        <span className="text-[26px] font-bold tracking-[-0.01em]">How do you eat?</span>
                        <span className="text-sm text-stone-500">
                            Help us tailor your meal plan to your needs. You can change this later in Profile.
                        </span>
                    </div>

                    <PreferencesForm
                        value={preferences}
                        onChange={setPreferences}
                        variant="onboarding"
                    />

                    <div className="flex flex-col gap-1.5">
                        <button
                            onClick={handleSubmit}
                            disabled={loading}
                            className="w-full bg-orange-500 hover:bg-orange-600 text-white text-[15px] font-semibold py-[13px] rounded-xl transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                        >
                            {loading ? 'Saving…' : 'Continue to my meal plan'}
                        </button>

                        <button
                            onClick={() => navigate('/meal-plan')}
                            className="w-full text-sm text-stone-500 hover:text-stone-900 py-2.5 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300 rounded-lg"
                        >
                            Skip for now
                        </button>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default DietaryPreferences
