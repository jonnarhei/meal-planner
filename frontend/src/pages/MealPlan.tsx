import { useEffect, useState } from "react"
import type { UserRecipe, MealPlan } from "../api/types"
import { changeRecipeForDay, getCurrentMealPlan, regenerateMealPlan } from "../api/mealplan"
import HamburgerMenu from "./HamburgerMenu"
import toast from "react-hot-toast"
import { getRecipes } from "../api/recipes"
import { useNavigate } from "react-router-dom"
import ChangeRecipePopover from "../components/ChangeRecipePopover"

function MealPlanPage() {
    const [mealPlan, setMealPlan] = useState<MealPlan | null>(null)
    const [loading, setLoading] = useState(true)
    const [changingDay, setChangingDay] = useState<number | null>(null)
    const [regenerating, setRegenerating] = useState(false)
    const [pickerDay, setPickerDay] = useState<number | null>(null)
    const [myRecipes, setMyRecipes] = useState<UserRecipe[] | null>(null)
    const [loadingRecipes, setLoadingRecipes] = useState(false)

    const navigate = useNavigate()

    useEffect(() => {
        const fetchMealPlan = async () => {
            try {
                const data = await getCurrentMealPlan()
                setMealPlan(data)
            } catch (err) {
                toast.error("Error fetching current meal plan, try again later")
            } finally {
                setLoading(false)
            }
        }

        fetchMealPlan()
    }, [])


    const handleRecipeChange = async (day: number, userRecipeId?: number) => {
        setPickerDay(null)
        setChangingDay(day)
        try {
            const updatedRecipe = await changeRecipeForDay(day, userRecipeId)
            setMealPlan(prev => ({
                ...prev!,
                recipes: prev!.recipes.map(r =>
                    r.day === day ? updatedRecipe : r
                )
            }))
            toast.success('Recipe changed! Regenerate your shopping list to update ingredients')
        } catch (err) {
            toast.error('Failed to change recipe')
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
            toast.error('Failed to regenerate the meal plan')
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

    const openPicker = async (day: number) => {
        setPickerDay(day)
        if (myRecipes !== null) return

        setLoadingRecipes(true)
        try {
            setMyRecipes(await getRecipes() ?? [])
        } catch (err) {
            toast.error('Failed to load your recieps')
            setMyRecipes([])
        } finally {
            setLoadingRecipes(false)
        }
    }

    const dayNames = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500 text-lg font-medium">Loading your meal plan...</p>
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
                <h2 className="text-xl font-semibold text-gray-700 mb-6">This Week's Meals</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-6">
                    {mealPlan.recipes.map((recipe) => (
                        <div key={recipe.day} className="bg-white rounded-3xl shadow-md border border-orange-100 flex flex-col">
                            {recipe.image ? (
                                <img
                                    src={recipe.image}
                                    alt={recipe.recipe_title}
                                    className="w-full h-40 object-cover object-center rounded-t-3xl"
                                />
                            ) : (
                                <div className="w-full h-40 bg-orange-100 flex items-center justify-center text-orange-300 text-4xl rounded-t-3xl">
                                    🍽
                                </div>
                            )}
                            <div className="p-4 flex flex-col flex-1">
                                <div>
                                    <p className="text-xs font-semibold text-orange-400 uppercase tracking-wide mb-1">
                                        {dayNames[recipe.day - 1]}
                                    </p>
                                    {recipe.source === 'user' && (
                                        <span className="text-xs bg-orange-100 text-orange-400 px-2 py-0.5 rounded-full">
                                            my recipe
                                        </span>
                                    )}
                                </div>
                                <h3 className="text-gray-800 font-semibold text-sm mb-3 leading-snug flex-1">
                                    {recipe.recipe_title}
                                </h3>
                                <div className="flex items-center justify-between">
                                    {recipe.source_url ? (
                                        <a
                                            href={recipe.source_url}
                                            target="_blank"
                                            rel="noreferrer"
                                            className="text-xs text-orange-500 hover:underline font-medium"
                                        >
                                            View Recipe
                                        </a>
                                    ) : <span />}
                                    <ChangeRecipePopover
                                        busy={changingDay === recipe.day}
                                        onSurprise={() => handleRecipeChange(recipe.day)}
                                        onUseOwn={() => openPicker(recipe.day)}
                                    />
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {pickerDay !== null && (
                <div
                    className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 px-6"
                    onClick={() => setPickerDay(null)}
                >
                    <div
                        className="bg-white rounded-3xl shadow-lg border border-orange-100 w-full max-w-md max-h-[80vh] flex flex-col"
                        onClick={e => e.stopPropagation()}
                    >
                        <div className="flex items-center justify-between p-5 border-b border-orange-100">
                            <h3 className="font-semibold text-gray-700">
                                {dayNames[pickerDay - 1]}
                            </h3>
                            <button
                                onClick={() => setPickerDay(null)}
                                className="text-gray-300 hover:text-gray-500 text-xl transition-colors"
                            >
                                ×
                            </button>
                        </div>

                        <div className="p-5 overflow-y-auto">
                            {loadingRecipes ? (
                                <p className="text-sm text-gray-400 text-center py-4">Loading...</p>
                            ) : (myRecipes?.length ?? 0) === 0 ? (
                                <div className="text-center py-4">
                                    <p className="text-sm text-gray-400 mb-3">You have no recipes yet</p>
                                    <button
                                        onClick={() => navigate('/recipes/new')}
                                        className="text-sm bg-orange-100 hover:bg-orange-200 text-orange-600 font-medium px-4 py-2 rounded-xl transition-colors"
                                    >
                                        Add a recipe
                                    </button>
                                </div>
                            ) : (
                                <div className="flex flex-col gap-2">
                                    {myRecipes!.map(recipe => (
                                        <button
                                            key={recipe.id}
                                            onClick={() => handleRecipeChange(pickerDay, recipe.id)}
                                            className="text-left border border-gray-200 hover:border-orange-300 hover:bg-orange-50 rounded-xl px-4 py-3 transition-colors"
                                        >
                                            <span className="block text-sm font-medium text-gray-700">
                                                {recipe.title}
                                            </span>
                                            <span className="block text-xs text-gray-400">
                                                {recipe.ingredients?.length ?? 0} ingredients
                                            </span>
                                        </button>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>

    )
}

export default MealPlanPage