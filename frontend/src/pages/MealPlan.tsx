import { useEffect, useState } from "react"
import type { UserRecipe, MealPlan, MealPlanRecipe } from "../api/types"
import { changeRecipeForDay, getCurrentMealPlan, regenerateMealPlan } from "../api/mealplan"
import toast from "react-hot-toast"
import { getRecipes } from "../api/recipes"
import { useNavigate } from "react-router-dom"
import { useAppLayout } from "../components/AppLayout"
import RecipeThumb from "../components/RecipeThumb"
import EmptyState from "../components/EmptyState"

/** The API sends dates as RFC3339. Read the calendar day only, so the local timezone can't shift it. */
const parseApiDate = (value: string) => {
    const [year, month, day] = value.slice(0, 10).split('-').map(Number)
    return new Date(year, month - 1, day)
}

const addDays = (date: Date, days: number) => {
    const next = new Date(date)
    next.setDate(next.getDate() + days)
    return next
}

const startOfToday = () => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return today
}

const shortDay = (date: Date) => date.toLocaleDateString('en-GB', { weekday: 'short' })
const longDay = (date: Date) => date.toLocaleDateString('en-GB', { weekday: 'long' })
// en-US so September abbreviates to "Sep" rather than en-GB's "Sept".
const shortMonth = (date: Date) => date.toLocaleDateString('en-US', { month: 'short' })

const formatWeekLabel = (start: Date, end: Date) => {
    if (start.getFullYear() !== end.getFullYear()) {
        return `${start.getDate()} ${shortMonth(start)} ${start.getFullYear()} – ${end.getDate()} ${shortMonth(end)} ${end.getFullYear()}`
    }
    if (start.getMonth() !== end.getMonth()) {
        return `${start.getDate()} ${shortMonth(start)} – ${end.getDate()} ${shortMonth(end)} ${end.getFullYear()}`
    }
    return `${start.getDate()} – ${end.getDate()} ${shortMonth(end)} ${end.getFullYear()}`
}

const DAY_IN_MS = 24 * 60 * 60 * 1000

function MealPlanPage() {
    const [mealPlan, setMealPlan] = useState<MealPlan | null>(null)
    const [loading, setLoading] = useState(true)
    const [changingDay, setChangingDay] = useState<number | null>(null)
    const [regenerating, setRegenerating] = useState(false)
    const [pickerDay, setPickerDay] = useState<number | null>(null)
    const [myRecipes, setMyRecipes] = useState<UserRecipe[] | null>(null)
    const [loadingRecipes, setLoadingRecipes] = useState(false)
    const [selectedDay, setSelectedDay] = useState<number | null>(null)

    const navigate = useNavigate()
    const { setWeekLabel } = useAppLayout()

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

    // Loaded up front so the list and the detail panel can show ingredient and serving
    // counts for our own recipes. The picker reuses whatever is already here.
    useEffect(() => {
        const fetchRecipes = async () => {
            try {
                setMyRecipes(await getRecipes() ?? [])
            } catch (err) {
                setMyRecipes([])
            }
        }

        fetchRecipes()
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
            setSelectedDay(null)
        } catch (err) {
            toast.error('Failed to regenerate the meal plan')
        } finally {
            setRegenerating(false)
        }
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

    const startDate = mealPlan ? parseApiDate(mealPlan.start_date) : null
    const endDate = mealPlan ? parseApiDate(mealPlan.end_date) : null
    const weekLabel = startDate && endDate ? formatWeekLabel(startDate, endDate) : null

    useEffect(() => {
        setWeekLabel(weekLabel)
        return () => setWeekLabel(null)
    }, [weekLabel, setWeekLabel])

    const dateForDay = (day: number) => addDays(startDate!, day - 1)

    const recipes = [...(mealPlan?.recipes ?? [])].sort((a, b) => a.day - b.day)

    /** Today's day number when this week's plan covers it, otherwise Monday. */
    const todayDay = (() => {
        if (!startDate) return 1
        const offset = Math.round((startOfToday().getTime() - startDate.getTime()) / DAY_IN_MS)
        return offset >= 0 && offset < 7 ? offset + 1 : 1
    })()

    const activeDay = selectedDay ?? todayDay
    const selected = recipes.find(r => r.day === activeDay) ?? recipes[0] ?? null

    /** "3 ingredients - 4 servings" for recipes we own. Unknown for Spoonacular ones. */
    const metaFor = (recipe: MealPlanRecipe) => {
        if (recipe.source !== 'user' || recipe.user_recipe_id === null) return ''

        const mine = myRecipes?.find(r => r.id === recipe.user_recipe_id)
        if (!mine) return ''

        const count = mine.ingredients?.length ?? 0
        const ingredients = count === 0 ? 'No ingredients' : `${count} ingredient${count === 1 ? '' : 's'}`

        return mine.servings > 0 ? `${ingredients} · ${mine.servings} servings` : ingredients
    }

    const busy = selected !== null && changingDay === selected.day

    if (loading) return <MealPlanSkeleton />

    if (!mealPlan || recipes.length === 0) return (
        <EmptyState
            visual={
                <div className="grid grid-cols-7 gap-[5px]">
                    {Array.from({ length: 7 }, (_, i) => (
                        <span
                            key={i}
                            className={`w-[22px] h-[30px] rounded-md ${i === 3
                                ? 'bg-orange-500'
                                : 'border-[1.5px] border-dashed border-orange-300'
                            }`}
                        />
                    ))}
                </div>
            }
            title="No dinners planned this week"
            description="We'll fill all seven days using your food preferences and recipes. You can change any day afterwards."
            actionLabel={regenerating ? 'Generating…' : 'Generate meal plan'}
            onAction={handleRegenerate}
            busy={regenerating}
        />
    )

    return (
        <>
            <div className="grid grid-cols-[400px_minmax(0,1fr)] gap-7 items-start">

                {/* Week list */}
                <div className="flex flex-col gap-1">
                    <div className="flex items-center justify-between px-1 pb-2.5">
                        <span className="text-[13px] font-semibold text-stone-500">
                            {recipes.length} dinner{recipes.length === 1 ? '' : 's'}
                        </span>
                        <button
                            onClick={handleRegenerate}
                            disabled={regenerating}
                            className="text-[13px] font-semibold text-orange-700 py-1 rounded transition-colors hover:text-orange-800 disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                        >
                            {regenerating ? 'Generating…' : 'New meal plan'}
                        </button>
                    </div>

                    {recipes.map(recipe => {
                        const date = dateForDay(recipe.day)
                        const isToday = recipe.day === todayDay
                        const isSelected = selected?.day === recipe.day
                        const meta = metaFor(recipe)

                        return (
                            <button
                                key={recipe.day}
                                onClick={() => setSelectedDay(recipe.day)}
                                className={`grid grid-cols-[44px_40px_minmax(0,1fr)] gap-3.5 items-center text-left px-3 py-2.5 rounded-xl border transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300
                                    ${isSelected
                                        ? 'bg-orange-50 border-orange-300'
                                        : 'border-transparent hover:bg-orange-50'
                                    }`}
                            >
                                <div className="flex flex-col items-center">
                                    <span className={`text-[11px] font-bold uppercase tracking-[0.06em] ${isToday ? 'text-orange-600' : 'text-stone-400'}`}>
                                        {shortDay(date)}
                                    </span>
                                    <span className="text-lg font-bold text-stone-900">{date.getDate()}</span>
                                </div>

                                <RecipeThumb
                                    src={recipe.image}
                                    alt=""
                                    className="w-10 h-10 rounded-[9px]"
                                />

                                <div className="flex flex-col gap-0.5 min-w-0">
                                    <span className="text-[15px] font-semibold text-stone-900 truncate">
                                        {recipe.recipe_title}
                                    </span>
                                    {meta && <span className="text-xs text-stone-500">{meta}</span>}
                                </div>
                            </button>
                        )
                    })}
                </div>

                {/* Detail panel */}
                {selected && (
                    <div className="border border-orange-200 rounded-[20px] overflow-hidden bg-white">
                        <RecipeThumb
                            src={selected.image}
                            alt={selected.recipe_title}
                            stripe={10}
                            className="block w-full h-[260px]"
                        />

                        <div className="px-7 pt-6 pb-7 flex flex-col gap-3.5">
                            <span className="text-[13px] font-bold uppercase tracking-[0.06em] text-orange-500">
                                {longDay(dateForDay(selected.day))} · {dateForDay(selected.day).getDate()} {shortMonth(dateForDay(selected.day))}
                            </span>

                            {busy ? (
                                <div className="h-8 w-3/4 rounded-md bg-stone-100" />
                            ) : (
                                <h2 className="text-[30px] font-bold tracking-[-0.01em] leading-[1.15]">
                                    {selected.recipe_title}
                                </h2>
                            )}

                            {(selected.source === 'user' || metaFor(selected)) && (
                                <div className="flex items-center gap-2.5 text-sm text-stone-600">
                                    {selected.source === 'user' && (
                                        <span className="bg-orange-100 text-orange-700 rounded-full px-2.5 py-0.5 text-[13px] font-medium">
                                            my recipe
                                        </span>
                                    )}
                                    <span>{metaFor(selected)}</span>
                                </div>
                            )}

                            {selected.source_url && (
                                <a
                                    href={selected.source_url}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="w-fit text-sm font-semibold text-orange-700 hover:text-orange-800 transition-colors"
                                >
                                    View recipe
                                </a>
                            )}

                            <div className="h-px bg-stone-100 my-1.5" />

                            <span className="text-[13px] text-stone-500">Change this dinner</span>

                            <div className="flex gap-2.5">
                                <button
                                    onClick={() => handleRecipeChange(selected.day)}
                                    disabled={busy}
                                    className={`bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-[18px] py-2.5 rounded-[10px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300 ${busy ? 'opacity-55' : ''}`}
                                >
                                    {busy ? 'Finding a dinner…' : 'Surprise me'}
                                </button>
                                <button
                                    onClick={() => openPicker(selected.day)}
                                    disabled={busy}
                                    className={`bg-orange-100 hover:bg-orange-200 text-orange-700 text-sm font-semibold px-[18px] py-2.5 rounded-[10px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300 ${busy ? 'opacity-55' : ''}`}
                                >
                                    Use my own
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {pickerDay !== null && (
                <div
                    className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 px-6"
                    onClick={() => setPickerDay(null)}
                >
                    <div
                        className="bg-white rounded-[20px] shadow-lg border border-orange-200 w-full max-w-md max-h-[80vh] flex flex-col"
                        onClick={e => e.stopPropagation()}
                    >
                        <div className="flex items-center justify-between px-5 py-4 border-b border-stone-100">
                            <h3 className="text-[15px] font-semibold text-stone-900">
                                {startDate ? longDay(dateForDay(pickerDay)) : 'Pick your recipe'}
                            </h3>
                            <button
                                onClick={() => setPickerDay(null)}
                                aria-label="Close"
                                className="text-stone-300 hover:text-red-600 text-xl leading-none transition-colors"
                            >
                                ×
                            </button>
                        </div>

                        <div className="p-5 overflow-y-auto">
                            {loadingRecipes ? (
                                <p className="text-sm text-stone-400 text-center py-4">Loading...</p>
                            ) : (myRecipes?.length ?? 0) === 0 ? (
                                <div className="text-center py-4">
                                    <p className="text-sm text-stone-400 mb-3">You have no recipes yet</p>
                                    <button
                                        onClick={() => navigate('/recipes/new')}
                                        className="bg-orange-100 hover:bg-orange-200 text-orange-700 text-sm font-semibold px-4 py-2 rounded-[10px] transition-colors"
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
                                            className="text-left border border-stone-200 hover:border-orange-300 hover:bg-orange-50 rounded-xl px-4 py-3 transition-colors"
                                        >
                                            <span className="block text-sm font-semibold text-stone-900">
                                                {recipe.title}
                                            </span>
                                            <span className="block text-xs text-stone-500">
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
        </>
    )
}

const skeletonWidths = ['70%', '55%', '80%', '60%', '75%', '50%']

function MealPlanSkeleton() {
    return (
        <div className="grid grid-cols-[400px_minmax(0,1fr)] gap-7 items-start animate-pulse">
            <div className="flex flex-col gap-2.5">
                <div className="h-3 w-[70px] rounded-md bg-stone-100 mx-1 mb-2" />

                {skeletonWidths.map((width, i) => (
                    <div key={i} className="flex items-center gap-2.5 p-1.5">
                        <div className="w-8 h-8 rounded-lg bg-stone-100" />
                        <div className="flex-1 flex flex-col gap-1.5">
                            <div className="h-[11px] rounded-md bg-stone-100" style={{ width }} />
                            <div className="h-[9px] w-[60px] rounded-md bg-stone-50" />
                        </div>
                    </div>
                ))}
            </div>

            <div className="border border-stone-100 rounded-2xl overflow-hidden flex flex-col">
                <div className="h-[140px] bg-stone-50" />
                <div className="p-[18px] flex flex-col gap-2.5">
                    <div className="h-2.5 w-[90px] rounded-md bg-orange-100" />
                    <div className="h-5 w-[70%] rounded-md bg-stone-100" />
                    <div className="h-[11px] w-[50%] rounded-md bg-stone-50" />
                </div>
            </div>
        </div>
    )
}

export default MealPlanPage
