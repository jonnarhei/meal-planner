import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { createRecipe, getRecipe, updateRecipe } from "../api/recipes";
import toast from "react-hot-toast";
import type { RecipeInput } from "../api/types";

type IngredientRow = { name: string; amount: string; unit: string; }

const emptyRow: IngredientRow = { name: '', amount: '', unit: '' }

function RecipeForm() {
    const { id } = useParams()
    const isEdit = id !== undefined

    const [form, setForm] = useState({ title: '', image: '', source_url: '', instructions: '', servings: '' })
    const [ingredients, setIngredients] = useState<IngredientRow[]>([emptyRow])
    const [loading, setLoading] = useState(isEdit)
    const [saving, setSaving] = useState(false)

    const navigate = useNavigate()

    useEffect(() => {
        if (!isEdit) return

        const fetchRecipes = async () => {
            try {
                const recipe = await getRecipe(Number(id))
                setForm({
                    title: recipe.title,
                    image: recipe.image,
                    source_url: recipe.source_url,
                    instructions: recipe.instructions,
                    servings: recipe.servings > 0 ? String(recipe.servings) : '',
                })

                const rows = (recipe.ingredients ?? []).map(ing => ({
                    name: ing.name,
                    amount: ing.amount > 0 ? String(ing.amount) : '',
                    unit: ing.unit,
                }))
                setIngredients(rows.length > 0 ? rows : [emptyRow])
            } catch (err) {
                toast.error('Could not load that recipe')
            } finally {
                setLoading(false)
            }
        }

        fetchRecipes()
    }, [id, isEdit, navigate])

    const updateField = (field: keyof typeof form, value: string) =>
        setForm(prev => ({ ...prev, [field]: value }))

    const updateIngredient = (index: number, field: keyof IngredientRow, value: string) =>
        setIngredients(prev => prev.map((row, i) => i === index ? { ...row, [field]: value } : row))

    const addRow = () => setIngredients(prev => [...prev, emptyRow])

    const removeRow = (index: number) =>
        setIngredients(prev => prev.length === 1 ? [emptyRow] : prev.filter((_, i) => i !== index))

    const handleSubmit = async () => {
        if (!form.title.trim()) {
            toast.error('Title is required')
            return
        }

        const filled = ingredients.filter(row => row.name.trim() !== '')

        const payload: RecipeInput = {
            title: form.title.trim(),
            image: form.image.trim(),
            source_url: form.source_url.trim(),
            instructions: form.instructions,
            servings: parseInt(form.servings) || 0,
            ingredients: filled.map(row => ({
                name: row.name.trim(),
                amount: parseFloat(row.amount) || 0,
                unit: row.unit.trim(),
            })),
        }

        setSaving(true)
        try {
            if (isEdit) {
                await updateRecipe(Number(id), payload)
                toast.success('Recipe updated')
            } else {
                await createRecipe(payload)
                toast.success('Recipe added')
            }
            navigate('/recipes')
        } catch (err: any) {
            toast.error(err.response?.data?.error ?? 'Failed to save recipe')
        } finally {
            setSaving(false)
        }
    }

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500 text-lg font-medium">Loading recipe...</p>
        </div>
    )

    const inputClasses = "border border-gray-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"

    return (
        <div className="min-h-screen bg-orange-50 py-8 px-6">
            <div className="bg-white rounded-3xl shadow-md border border-orange-100 p-8 w-full max-w-2xl mx-auto">
                <h1 className="text-2xl font-bold text-orange-600 mb-2">
                    {isEdit ? 'Edit Recipe' : 'New Recipe'}
                </h1>
                <p className="text-gray-500 text-sm mb-6">
                    Only a title is required. Add ingredients if you want them on your shopping list.
                </p>

                <label className="block text-sm font-semibold text-gray-700 mb-2">Title</label>
                <input
                    type="text"
                    placeholder="e.g. Grandma's lasagna"
                    value={form.title}
                    onChange={e => updateField('title', e.target.value)}
                    className={`${inputClasses} w-full mb-4`}
                />

                <div className="flex gap-3 mb-4">
                    <div className="flex-1">
                        <label className="block text-sm font-semibold text-gray-700 mb-2">Image URL</label>
                        <input
                            type="text"
                            placeholder="https://... (optional)"
                            value={form.image}
                            onChange={e => updateField('image', e.target.value)}
                            className={`${inputClasses} w-full`}
                        />
                    </div>
                    <div className="w-28">
                        <label className="block text-sm font-semibold text-gray-700 mb-2">Servings</label>
                        <input
                            type="number"
                            min="0"
                            placeholder="4"
                            value={form.servings}
                            onChange={e => updateField('servings', e.target.value)}
                            className={`${inputClasses} w-full`}
                        />
                    </div>
                </div>

                <label className="block text-sm font-semibold text-gray-700 mb-2">Source link</label>
                <input
                    type="text"
                    placeholder="https://... (optional)"
                    value={form.source_url}
                    onChange={e => updateField('source_url', e.target.value)}
                    className={`${inputClasses} w-full mb-6`}
                />

                <div className="flex items-center justify-between mb-3">
                    <label className="text-sm font-semibold text-gray-700">Ingredients</label>
                    <button
                        onClick={addRow}
                        className="text-xs bg-orange-100 hover:bg-orange-200 text-orange-600 font-medium px-3 py-1.5 rounded-lg transition-colors"
                    >
                        Add ingredient
                    </button>
                </div>

                <div className="flex flex-col gap-2 mb-6">
                    {ingredients.map((row, index) => (
                        <div key={index} className="flex gap-2 items-center">
                            <input
                                type="text"
                                placeholder="Ingredient"
                                value={row.name}
                                onChange={e => updateIngredient(index, 'name', e.target.value)}
                                className={`${inputClasses} flex-1 min-w-0`}
                            />
                            <input
                                type="number"
                                min="0"
                                step="any"
                                placeholder="Amount"
                                value={row.amount}
                                onChange={e => updateIngredient(index, 'amount', e.target.value)}
                                className={`${inputClasses} w-24 shrink-0`}
                            />
                            <input
                                type="text"
                                placeholder="Unit"
                                value={row.unit}
                                onChange={e => updateIngredient(index, 'unit', e.target.value)}
                                className={`${inputClasses} w-24 shrink-0`}
                            />
                            <button
                                onClick={() => removeRow(index)}
                                className="text-gray-300 hover:text-red-400 transition-colors text-lg px-1"
                            >
                                ×
                            </button>
                        </div>
                    ))}
                </div>

                <label className="block text-sm font-semibold text-gray-700 mb-2">Instruction</label>
                <textarea
                    rows={6}
                    placeholder="Optional. How do you make it?"
                    value={form.instructions}
                    onChange={e => updateField('instructions', e.target.value)}
                    className={`${inputClasses} w-full mb-6 resize-y`}
                />

                <button
                    onClick={handleSubmit}
                    disabled={saving}
                    className="w-full bg-orange-500 hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition-colors disabled:opacity-50"
                >
                    {saving ? 'Saving...' : isEdit ? 'Save changes' : 'Add recipe'}
                </button>

                <button
                    onClick={() => navigate('/recipes')}
                    className="w-full text-gray-400 hover:text-gray-600 text-sm mt-3 py-2 transition-colors"
                >
                    Cancel
                </button>
            </div>
        </div>
    )
}

export default RecipeForm