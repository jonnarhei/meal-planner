import { useEffect, useState } from "react"
import type { MealPlan } from "../api/types"
import { changeRecipeForDay, getCurrentMealPlan, regenerateMealPlan } from "../api/mealplan"
import HamburgerMenu from "./HamburgerMenu"

function MealPlanPage() {
    const [mealPlan, setMealPlan] = useState<MealPlan | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [changingDay, setChangingDay] = useState<number | null>(null)
    const [regenerating, setRegenerating] = useState(false)

    useEffect(() => {
        const fetchMealPlan = async () => {
            try {
                const data = await getCurrentMealPlan()
                setMealPlan(data)
            } catch (err) {
                setError("Error fetching current meal plan, try again later")
            } finally {
                setLoading(false)
            }
        }

        fetchMealPlan()
    }, [])


    const handleRecipeChange = async (day: number) => {
        setChangingDay(day)
        try {
            const updatedRecipe = await changeRecipeForDay(day)
            setMealPlan(prev => ({
                ...prev!,
                recipes: prev!.recipes.map(r => 
                    r.day === day ? updatedRecipe : r
                )
            }))
        } catch (err) {
            setError('Failed to change recipe')
        } finally {
            setChangingDay(null)
        }
    }

    const handleRegenerate = async () => {
        setRegenerating(true)

        try {
            const newMealPlan = await regenerateMealPlan()
            setMealPlan(newMealPlan)
        } catch (err) {
            setError('Failed to regenerate the meal plan')
        } finally {
            setRegenerating(false)
        }
    }

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString('en-GB', {
            month: 'long',
            day: 'numeric',
            year: 'numeric'
        })
    }

    const dayNames = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500 text-lg font-medium">Loading your meal plan...</p>
        </div>
    )

    if (error) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-red-500 text-lg">{error}</p>
        </div>
    )

    if (!mealPlan) return null

    return (
        <div className="min-h-screen bg-orange-50">

            <div className="bg-white shadow-sm">
                <div className="max-w-screen-2xl mx-auto px-6 py-4 flex justify-between items-center">
                    <div className="flex items-center gap-4">
                        <HamburgerMenu />
                        <h1 className="text-2xl font-bold text-orange-600">Meal Planner</h1>
                    </div>
                    <p className="text-sm text-gray-500">
                        {formatDate(mealPlan.start_date)} - {formatDate(mealPlan.end_date)}
                    </p>
                    <button 
                        onClick={handleRegenerate}
                        disabled={regenerating}
                        className="bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-4 py-2 rounded-xl transition-colors disabled:opacity-50"
                    >
                        {regenerating ? 'Generating' : 'New Meal Plan'}
                    </button>
                </div>
            </div>

            <div className="max-w-screen-2xl mx-auto px-6 py-8">
                <h2 className="text-x1 font-semibold text-gray-700 mb-6">This Week's Meals</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-6">
                    {mealPlan.recipes.sort((a, b) => a.day - b.day).map((recipe) => (
                        <div key={recipe.day} className="bg-white rounded-3xl shadow-md border border-orange-100 overflow-hidden">
                            <img 
                                src={recipe.image} 
                                alt={recipe.recipe_title}
                                className="w-full h-40 object-cover object-center" 
                            />
                            <div className="p-4 flex flex-col flex-1">
                                <p className="text-xs font-semibold text-orange-400 uppercase tracking-wide mb-1">
                                    {dayNames[recipe.day - 1]}
                                </p>
                                <h3 className="text-gray-800 font-semibold text-sm mb-3 leading-snug flex-1">
                                    {recipe.recipe_title}
                                </h3>
                                <div className="flex items-center justify-between">
                                    <a 
                                        href={recipe.source_url} 
                                        target="_blank"
                                        rel="norefferer"
                                        className="text-xs text-orange-500 hover-underline fond-medium"
                                    >
                                        View Recipe
                                    </a>
                                    <button 
                                        onClick={() => handleRecipeChange(recipe.day)}
                                        disabled={changingDay === recipe.day}
                                        className="text-cs bg-orange-100 hover:bg-orange-200 test-orange-600 font-medium px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50"
                                    >
                                        {changingDay === recipe.day ? 'Changing...' : 'Change'}
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>

    )
}

export default MealPlanPage