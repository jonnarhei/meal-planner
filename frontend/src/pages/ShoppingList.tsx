import { useEffect, useState } from "react";
import type { ShoppingListItem } from "../api/types";
import { addFromMealPlan, addItems, deleteChecked, deleteItem, getShoppingList, toggleChecked } from "../api/shoppinglist";
import toast from "react-hot-toast";
import HamburgerMenu from "./HamburgerMenu";

function ShoppingList() {
    const [items, setItems] = useState<ShoppingListItem[]>([])
    const [loading, setLoading] = useState(true)
    const [addingFromPlan, setAddingFromPlan] = useState(false)
    const [newItem, setNewItem] = useState({ name: '', amount: '', unit: '' })

    useEffect(() => {
        const fetchItems = async () => {
            try {
                const data = await getShoppingList()
                setItems(data ?? [])
            } catch (err) {
                toast.error('Failed to load shopping list')
            } finally {
                setLoading(false)
            }
        }

        fetchItems()
    }, [])

    const handleAddFromMealPlan = async () => {
        setAddingFromPlan(true)
        try {
            await addFromMealPlan()
            const data = await getShoppingList()
            setItems(data ?? [])
            toast.success('Ingredients added from meal plan!')
        } catch (error) {
            toast.error('Failed to add ingredients')
        } finally {
            setAddingFromPlan(false)
        }
    }

    const handleToggle = async (id: number) => {
        try {
            await toggleChecked(id)
            setItems(prev => prev.map(item => 
                item.id === id ? { ...item, checked: !item.checked } : item
            ))
        } catch (error) {
            toast.error('Failed to update item')
        }
    }

    const handleDelete = async (id:number) => {
        try {
            await deleteItem(id)
            setItems(prev => prev.filter(item => item.id !== id))
        } catch (error) {
            toast.error('Failed to delete item')
        }
    }


    const handleDeleteChecked = async () => {
        try {
            await deleteChecked()
            setItems(prev => prev.filter(items => !items.checked))
            toast.success('Checked items cleared')
        } catch (error) {
            toast.error('Failed to clear checked items')
        }
    }

    const handleAddItem = async () => {
        if (!newItem.name) {
            toast.error('Item name is required')
            return
        }
        try {
            await addItems([{
                name: newItem.name,
                amount: parseFloat(newItem.amount) || 0,
                unit: newItem.unit
            }])
            const data = await getShoppingList()
            setItems(data ?? [])
            setNewItem({ name: '', amount: '', unit: '' })
            toast.success('Item added!')
        } catch (error) {
            toast.error('Failed to add item')
        }
    }

    const checkedCount = items.filter(i => i.checked).length

    const roundAmount = (amount: number, unit: string): number => {
        if (amount === 0) return 0

        //round to nearest 10 for grams and milliliters
        if (unit === 'grams' || unit === 'milliliters' || unit === 'g' || unit === 'ml') {
            return Math.round(amount / 10) * 10
        }

        //round to 1 decimal for everything else
        return Math.round(amount * 10) / 10
    }

    if (loading) return (
        <div className="min-h-screen bg-orange-50 flex items-center justify-center">
            <p className="text-orange-500">Loading...</p>
        </div>
    )

    return (
        <div className="min-h-screen bg-orange-50">
            <div className="bg-white shadow-sm">
                <div className="max-w-screen-2xl mx-auto px-6 py-4 flex justify-between items-center">
                    <div className="flex items-center gap-4">
                        <HamburgerMenu />
                        <h1 className="text-2xl font-bold text-orange-600">Shopping List</h1>
                    </div>
                    <button
                        onClick={handleAddFromMealPlan}
                        disabled={addingFromPlan}
                        className="bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-4 py-2 rounded-xl transition-colors disabled:opacity-50"
                    >
                        {addingFromPlan ? 'Adding...' : '+ From Meal Plan'}
                    </button>
                </div>
            </div>

            <div className="max-w-2xl mx-auto px-6 py-8">

                {/* Add item manually */}
                <div className="bg-white rounded-3xl shadow-sm border border-orange-100 p-4 mb-6 flex gap-2">
                    <input
                        type="text"
                        placeholder="Item name"
                        value={newItem.name}
                        onChange={e => setNewItem(prev => ({ ...prev, name: e.target.value }))}
                        className="flex-1 border border-gray-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                    />
                    <input
                        type="number"
                        placeholder="Amount"
                        value={newItem.amount}
                        onChange={e => setNewItem(prev => ({ ...prev, amount: e.target.value }))}
                        className="w-24 border border-gray-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                    />
                    <input
                        type="text"
                        placeholder="Unit"
                        value={newItem.unit}
                        onChange={e => setNewItem(prev => ({ ...prev, unit: e.target.value }))}
                        className="w-24 border border-gray-200 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
                    />
                    <button
                        onClick={handleAddItem}
                        className="bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-4 py-2 rounded-xl transition-colors"
                    >
                        Add
                    </button>
                </div>

                {/* Clear checked button */}
                {checkedCount > 0 && (
                    <div className="flex justify-end mb-4">
                        <button
                            onClick={handleDeleteChecked}
                            className="text-sm text-red-400 hover:text-red-600 transition-colors"
                        >
                            Clear {checkedCount} checked item{checkedCount > 1 ? 's' : ''}
                        </button>
                    </div>
                )}

                {/* Shopping list items */}
                {items.length === 0 ? (
                    <div className="text-center py-12 text-gray-400">
                        <p className="text-lg mb-2">Your shopping list is empty</p>
                        <p className="text-sm">Add items manually or import from your meal plan</p>
                    </div>
                ) : (
                    <div className="flex flex-col gap-2">
                        {[...items]
                            .sort((a, b) => {
                                if (a.checked !== b.checked) return a.checked ? 1 : -1
                                return a.name.localeCompare(b.name)
                            })
                            .map(item => (
                            <div
                                key={item.id}
                                className="bg-white rounded-2xl shadow-sm border border-orange-100 px-4 py-3 flex items-center gap-3"
                            >
                                <input
                                    type="checkbox"
                                    checked={item.checked}
                                    onChange={() => handleToggle(item.id)}
                                    className="w-4 h-4 accent-orange-500"
                                />
                                <div className="flex-1">
                                    <span className={`text-sm font-medium capitalize ${item.checked ? 'line-through text-gray-400' : 'text-gray-700'}`}>
                                        {item.name}
                                    </span>
                                    {(item.amount > 0 || item.unit) && (
                                        <span className="text-xs text-gray-400 ml-2">
                                            {item.amount > 0 ? roundAmount(item.amount, item.unit) : ''} {item.unit}
                                        </span>
                                    )}
                                </div>
                                <span className={`text-xs px-2 py-0.5 rounded-full ${item.source === 'meal_plan' ? 'bg-orange-100 text-orange-400' : 'bg-gray-100 text-gray-400'}`}>
                                    {item.source === 'meal_plan' ? 'meal plan' : 'manual'}
                                </span>
                                <button
                                    onClick={() => handleDelete(item.id)}
                                    className="text-gray-300 hover:text-red-400 transition-colors text-lg"
                                >
                                    ×
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
}

export default ShoppingList