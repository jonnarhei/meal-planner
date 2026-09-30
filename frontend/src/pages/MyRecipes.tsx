import { useEffect, useState } from "react";
import type { UserRecipe } from "../api/types";
import { useNavigate } from "react-router-dom";
import { deleteRecipe, getRecipes } from "../api/recipes";
import toast from "react-hot-toast";
import RecipeThumb from "../components/RecipeThumb";
import EmptyState from "../components/EmptyState";
import MobileHeader from "../components/MobileHeader";

const columns = "grid grid-cols-[minmax(0,1fr)_140px_110px_150px] gap-4 px-5"

function MyRecipes() {
    const [recipes, setRecipes] = useState<UserRecipe[]>([])
    const [loading, setLoading] = useState(true)
    const [deletingId, setDeletingId] = useState<number | null>(null)
    const [query, setQuery] = useState('')

    const navigate = useNavigate()

    useEffect(() => {
        const fetchRecipes = async () => {
            try {
                const data = await getRecipes()
                setRecipes(data ?? [])
            } catch (err) {
                toast.error('Failed to load your recipes')
            } finally {
                setLoading(false)
            }
        }

        fetchRecipes()
    }, [])

    const handleDelete = async (recipe: UserRecipe) => {
        if (!window.confirm(`Delete "${recipe.title}"?`)) return

        setDeletingId(recipe.id)
        try {
            await deleteRecipe(recipe.id)
            setRecipes(prev => prev.filter(r => r.id !== recipe.id))
            toast.success('Recipe deleted')
        } catch (err) {
            toast.error('Failed to delete recipe')
        } finally {
            setDeletingId(null)
        }
    }

    if (loading) return (
        <>
            <MobileHeader title="Recipes" />
            <RecipesSkeleton />
        </>
    )

    if (recipes.length === 0) return (
        <>
        <MobileHeader title="Recipes" />
        <EmptyState
            visual={
                <RecipeThumb stripe={6} className="w-[72px] h-[72px] rounded-2xl" />
            }
            title="Your recipe book is empty"
            description={'Add the dinners you already make. They can go into your plan with "Use my own".'}
            actionLabel="Add your first recipe"
            onAction={() => navigate('/recipes/new')}
        />
        </>
    )

    const visible = recipes.filter(recipe =>
        recipe.title.toLowerCase().includes(query.trim().toLowerCase())
    )

    const noMatches = (
        <p className="px-1 md:px-5 py-6 text-sm text-stone-500">
            No recipes match "{query.trim()}".
        </p>
    )

    return (
        <div className="max-w-[920px] flex flex-col gap-4">
            <MobileHeader title="Recipes" />

            <div className="flex gap-2.5">
                <input
                    type="text"
                    placeholder="Search recipes"
                    value={query}
                    onChange={e => setQuery(e.target.value)}
                    className="flex-1 min-w-0 border border-stone-200 rounded-[10px] bg-stone-50 px-3.5 py-2.5 text-base md:text-[15px] outline-none transition-colors focus:border-orange-400"
                />
                <button
                    onClick={() => navigate('/recipes/new')}
                    className="flex-none min-h-11 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-[18px] rounded-[10px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                >
                    New recipe
                </button>
            </div>

            {/* Mobile: list rows instead of the table */}
            <div className="md:hidden flex flex-col">
                {visible.length === 0 ? noMatches : visible.map(recipe => {
                    const ingredientCount = recipe.ingredients?.length ?? 0
                    const meta = [
                        ingredientCount === 0 ? 'No ingredients' : `${ingredientCount} ingredient${ingredientCount === 1 ? '' : 's'}`,
                        recipe.servings > 0 && `${recipe.servings} servings`,
                    ].filter(Boolean).join(' · ')

                    return (
                        <div key={recipe.id} className="flex items-center gap-3 py-2.5 border-b border-stone-100">
                            <RecipeThumb
                                src={recipe.image}
                                alt=""
                                className="w-12 h-12 flex-none rounded-[10px]"
                            />
                            <div className="flex-1 min-w-0 flex flex-col gap-0.5">
                                <span className="text-[15px] font-semibold truncate">{recipe.title}</span>
                                <span className="text-xs text-stone-500">{meta}</span>
                            </div>
                            <div className="flex-none flex">
                                <button
                                    onClick={() => navigate(`/recipes/${recipe.id}/edit`)}
                                    className="min-h-11 text-sm font-semibold text-orange-700 active:bg-orange-50 rounded-lg px-2.5 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDelete(recipe)}
                                    disabled={deletingId === recipe.id}
                                    className="min-h-11 text-sm text-red-600 active:bg-red-50 rounded-lg px-2.5 transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                >
                                    {deletingId === recipe.id ? '...' : 'Delete'}
                                </button>
                            </div>
                        </div>
                    )
                })}
            </div>

            <div className="hidden md:block border border-stone-200 rounded-[14px] overflow-hidden">
                <div className={`${columns} py-2.5 bg-stone-50 text-xs font-semibold uppercase tracking-[0.04em] text-stone-500`}>
                    <span>Recipe</span>
                    <span>Ingredients</span>
                    <span>Servings</span>
                    <span />
                </div>

                {visible.length === 0 ? (
                    <p className="border-t border-stone-100 px-5 py-6 text-sm text-stone-500">
                        No recipes match "{query.trim()}".
                    </p>
                ) : visible.map(recipe => {
                    const ingredientCount = recipe.ingredients?.length ?? 0

                    return (
                        <div
                            key={recipe.id}
                            className={`${columns} py-2.5 items-center border-t border-stone-100`}
                        >
                            <div className="flex items-center gap-3 min-w-0">
                                <RecipeThumb
                                    src={recipe.image}
                                    alt=""
                                    className="w-10 h-10 flex-none rounded-[9px]"
                                />
                                <span className="text-[15px] font-semibold truncate">{recipe.title}</span>
                            </div>

                            <span className="text-sm text-stone-600">
                                {ingredientCount === 0 ? '—' : ingredientCount}
                            </span>
                            <span className="text-sm text-stone-600">
                                {recipe.servings > 0 ? recipe.servings : '—'}
                            </span>

                            <div className="flex justify-end gap-1">
                                <button
                                    onClick={() => navigate(`/recipes/${recipe.id}/edit`)}
                                    className="text-sm font-semibold text-orange-700 hover:bg-orange-50 rounded-lg px-2.5 py-1.5 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDelete(recipe)}
                                    disabled={deletingId === recipe.id}
                                    className="text-sm text-red-600 hover:bg-red-50 rounded-lg px-2.5 py-1.5 transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                >
                                    {deletingId === recipe.id ? '...' : 'Delete'}
                                </button>
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}

function RecipesSkeleton() {
    return (
        <div className="max-w-[920px] flex flex-col gap-4 animate-pulse">
            <div className="flex gap-2.5">
                <div className="flex-1 h-[42px] rounded-[10px] bg-stone-100" />
                <div className="w-[110px] h-[44px] rounded-[10px] bg-stone-100" />
            </div>

            <div className="md:hidden flex flex-col">
                {Array.from({ length: 5 }, (_, i) => (
                    <div key={i} className="flex items-center gap-3 py-2.5 border-b border-stone-100">
                        <div className="w-12 h-12 flex-none rounded-[10px] bg-stone-100" />
                        <div className="flex-1 flex flex-col gap-1.5">
                            <div className="h-[11px] w-[60%] rounded-md bg-stone-100" />
                            <div className="h-[9px] w-[40%] rounded-md bg-stone-50" />
                        </div>
                    </div>
                ))}
            </div>

            <div className="hidden md:block border border-stone-200 rounded-[14px] overflow-hidden">
                <div className="h-[38px] bg-stone-50" />
                {Array.from({ length: 5 }, (_, i) => (
                    <div key={i} className={`${columns} py-2.5 items-center border-t border-stone-100`}>
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 flex-none rounded-[9px] bg-stone-100" />
                            <div className="h-[11px] w-[55%] rounded-md bg-stone-100" />
                        </div>
                        <div className="h-[11px] w-8 rounded-md bg-stone-50" />
                        <div className="h-[11px] w-6 rounded-md bg-stone-50" />
                        <span />
                    </div>
                ))}
            </div>
        </div>
    )
}

export default MyRecipes
