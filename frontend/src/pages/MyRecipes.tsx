import { useEffect, useState } from "react";
import type { UserRecipe } from "../api/types";
import { useNavigate } from "react-router-dom";
import { deleteRecipe, getRecipes } from "../api/recipes";
import toast from "react-hot-toast";
import HamburgerMenu from "./HamburgerMenu";

function MyRecipes() {
    const [recipes, setRecipes] = useState<UserRecipe[]>([])
    const [loading, setLoading] = useState(true)
    const [deletingId, setDeletingId] = useState<number | null>(null)

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
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500 text-lg font-medium">Loading your recipes...</p>
        </div>
    )

    return (
        <div className="min-h-screen bg-orange-50">

            <div className="bg-white shadow-sm">
                <div className="max-w-screen-2xl mx-auto px-6 py-4 flex justify-between items-center">
                    <div className="flex items-center gap-4">
                        <HamburgerMenu />
                        <h1 className="text-2xl font-bold text-orange-600">My Recipes</h1>
                    </div>
                    <button
                        onClick={() => navigate('/recipes/new')}
                        className="bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-4 py-2 rounded-xl transition-colors"
                    >
                        New Recipe
                    </button>
                </div>
            </div>

            <div className="max-w-screen-2xl mx-auto px-6 py-8">
                {recipes.length === 0 ? (
                    <div className="text-center py-12 text-gray-400">
                        <p className="text-lg mb-2">You haven't added any recipes yet</p>
                        <p className="text-sm mb-6">Add your own recipes to use them in your meal plan</p>
                        <button
                            onClick={() => navigate('/recipes/new')}
                            className="bg-orange-100 hover:bg-orange-200 text-orange-600 text-sm font-semibold px-4 py-2 rounded-xl transition-colors"
                        >
                            Add your first recipe
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-6">
                        {recipes.map(recipe => {
                            const ingredientCount = recipe.ingredients?.length ?? 0

                            return (
                                <div
                                    key={recipe.id}
                                    className="bg-white rounded-3xl shadow-md border border-orange-100 overflow-hidden flex flex-col"
                                >
                                    {recipe.image ? (
                                        <img
                                            src={recipe.image}
                                            alt={recipe.title}
                                            onError={e => { e.currentTarget.style.display = 'none' }}
                                            className="w-full h-40 object-cover object-center"
                                        />
                                    ) : (
                                        <div className="w-full h-40 bg-orange-100 flex items-center justify-center text-orange-300 text-4xl">
                                            🍽
                                        </div>
                                    )}

                                    <div className="p-4 flex flex-col flex-1">
                                        <h3 className="text-gray-800 font-semibold text-sm mb-2 leading-snug flex-1">
                                            {recipe.title}
                                        </h3>

                                        <p className="text-xs text-gray-400 mb-3">
                                            {ingredientCount === 0
                                                ? 'No ingredients'
                                                : `${ingredientCount} ingredient${ingredientCount > 1 ? 's' : ''}`}
                                            {recipe.servings > 0 && ` ·${recipe.servings} servings`}
                                        </p>

                                        <div className="flex items-center justify-between">
                                            {recipe.source_url ? (
                                                <a
                                                    href={recipe.source_url}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className="text-xs text-orange-500 hover:underline font-medium"
                                                >
                                                    View Source
                                                </a>
                                            ) : <span />}

                                            <div className="flex gap-2">
                                                <button
                                                    onClick={() => navigate(`/recipes/${recipe.id}/edit`)}
                                                    className="text-xs bg-orange-100 hover:bg-orange-200 text-orange-600 font-medium px-3 py-1.5 rounded-lg transition-colors"
                                                >
                                                    Edit
                                                </button>
                                                <button
                                                    onClick={() => handleDelete(recipe)}
                                                    disabled={deletingId === recipe.id}
                                                    className="text-xs text-red-400 hover:text-red-600 font-medium px-2 py-1.5 transition-colors disabled:opacity-50"
                                                >
                                                    {deletingId === recipe.id ? '...' : 'Delete'}
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )
                        })}
                    </div>
                )}
            </div>
        </div>
    )
}

export default MyRecipes