import { useEffect, useState } from "react";
import type { ShoppingListItem } from "../api/types";
import { addFromMealPlan, addItems, deleteChecked, deleteItem, getShoppingList, toggleChecked } from "../api/shoppinglist";
import toast from "react-hot-toast";
import MobileHeader from "../components/MobileHeader";

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

    const handleDelete = async (id: number) => {
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

    const roundAmount = (amount: number, unit: string): number => {
        if (amount === 0) return 0

        //round to nearest 10 for grams and milliliters
        if (unit === 'grams' || unit === 'milliliters' || unit === 'g' || unit === 'ml') {
            return Math.round(amount / 10) * 10
        }

        //round to 1 decimal for everything else
        return Math.round(amount * 10) / 10
    }

    const quantity = (item: ShoppingListItem) => {
        if (item.amount <= 0 && !item.unit) return ''
        return `${item.amount > 0 ? roundAmount(item.amount, item.unit) : ''} ${item.unit}`.trim()
    }

    const byName = (a: ShoppingListItem, b: ShoppingListItem) => a.name.localeCompare(b.name)

    const toBuy = items.filter(item => !item.checked).sort(byName)
    const basket = items.filter(item => item.checked).sort(byName)

    if (loading) return (
        <>
            <MobileHeader title="Shopping list" action={null} />
            <ShoppingListSkeleton />
        </>
    )

    const nothingToBuy = (
        <div className="border-[1.5px] border-dashed border-orange-200 rounded-[14px] flex flex-col items-center justify-center gap-2 text-center p-8">
            <span className="text-[15px] font-semibold">Nothing left to buy</span>
            <span className="text-[13px] text-stone-500">Pull ingredients from this week's dinners.</span>
            <button
                onClick={handleAddFromMealPlan}
                disabled={addingFromPlan}
                className="mt-1 min-h-11 bg-orange-100 hover:bg-orange-200 text-orange-700 text-sm font-semibold px-4 py-2.5 rounded-[10px] transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
            >
                {addingFromPlan ? 'Updating…' : 'Update from meal plan'}
            </button>
        </div>
    )

    return (
        <>
        {/* Mobile (2j) */}
        <div className="md:hidden flex flex-col">
            <MobileHeader
                title="Shopping list"
                action={
                    <button
                        onClick={handleAddFromMealPlan}
                        disabled={addingFromPlan}
                        className="min-h-11 text-sm font-semibold text-orange-700 rounded transition-colors disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        {addingFromPlan ? 'Updating…' : 'Update'}
                    </button>
                }
            />

            <form
                onSubmit={e => { e.preventDefault(); handleAddItem() }}
                className="flex gap-1.5 border border-stone-200 rounded-xl p-[5px] mb-2"
            >
                <input
                    type="text"
                    placeholder="Add item"
                    value={newItem.name}
                    onChange={e => setNewItem(prev => ({ ...prev, name: e.target.value }))}
                    className="flex-1 min-w-0 border-none p-2.5 text-base outline-none"
                />
                <button
                    type="submit"
                    className="min-h-11 bg-orange-500 hover:bg-orange-600 text-white text-[15px] font-semibold px-[18px] rounded-lg transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                >
                    Add
                </button>
            </form>

            {toBuy.length === 0 ? nothingToBuy : toBuy.map(item => (
                <div key={item.id} className="flex items-center border-b border-stone-100">
                    <button
                        onClick={() => handleToggle(item.id)}
                        aria-label={`Tick off ${item.name}`}
                        className="flex-1 min-w-0 flex items-center gap-3.5 px-0.5 min-h-[52px] text-left rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        <span className="w-6 h-6 flex-none rounded-[7px] border-[1.5px] border-stone-300 bg-white" />
                        <span className="text-base capitalize truncate">{item.name}</span>
                        <span className="text-sm text-stone-500 flex-none">{quantity(item)}</span>
                    </button>
                    <button
                        onClick={() => handleDelete(item.id)}
                        aria-label={`Remove ${item.name}`}
                        className="w-11 h-11 -mr-2.5 flex-none flex items-center justify-center text-stone-300 hover:text-red-600 text-lg leading-none transition-colors"
                    >
                        ×
                    </button>
                </div>
            ))}

            <div className="flex items-center justify-between px-0.5 pt-[22px] pb-1.5">
                <span className="text-sm font-bold text-stone-600">In the basket · {basket.length}</span>
                {basket.length > 0 && (
                    <button
                        onClick={handleDeleteChecked}
                        className="min-h-11 text-sm text-red-600 rounded transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        Clear
                    </button>
                )}
            </div>

            {basket.length === 0 ? (
                <span className="px-0.5 text-[13px] text-stone-400">Items you tick off land here.</span>
            ) : basket.map(item => (
                <button
                    key={item.id}
                    onClick={() => handleToggle(item.id)}
                    aria-label={`Put ${item.name} back on the list`}
                    className="flex items-center gap-3.5 px-0.5 min-h-12 text-left rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                >
                    <span className="w-6 h-6 flex-none rounded-[7px] bg-orange-500 text-white text-[13px] font-bold flex items-center justify-center">
                        ✓
                    </span>
                    <span className="text-[15px] text-stone-400 line-through capitalize truncate">{item.name}</span>
                    <span className="text-[13px] text-stone-400 flex-none">{quantity(item)}</span>
                </button>
            ))}
        </div>

        {/* Desktop (1c) */}
        <div className="hidden md:grid grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)] gap-7 items-start">

            {/* To buy */}
            <div className="flex flex-col gap-3">
                <div className="flex items-center justify-between">
                    <span className="text-base font-bold">
                        To buy <span className="text-stone-400 font-medium">{toBuy.length}</span>
                    </span>
                    <button
                        onClick={handleAddFromMealPlan}
                        disabled={addingFromPlan}
                        className="text-[13px] font-semibold text-orange-700 rounded transition-colors hover:text-orange-800 disabled:opacity-55 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        {addingFromPlan ? 'Updating…' : 'Update from meal plan'}
                    </button>
                </div>

                <form
                    onSubmit={e => { e.preventDefault(); handleAddItem() }}
                    className="flex gap-1.5 border border-stone-200 rounded-xl p-[5px]"
                >
                    <input
                        type="text"
                        placeholder="Add item"
                        value={newItem.name}
                        onChange={e => setNewItem(prev => ({ ...prev, name: e.target.value }))}
                        className="flex-1 min-w-0 border-none px-2.5 py-2 text-[15px] outline-none"
                    />
                    <input
                        type="number"
                        min="0"
                        step="any"
                        placeholder="Amt"
                        value={newItem.amount}
                        onChange={e => setNewItem(prev => ({ ...prev, amount: e.target.value }))}
                        className="w-[60px] border-l border-stone-100 px-2.5 py-2 text-[15px] outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                    <input
                        type="text"
                        placeholder="Unit"
                        value={newItem.unit}
                        onChange={e => setNewItem(prev => ({ ...prev, unit: e.target.value }))}
                        className="w-[60px] border-l border-stone-100 px-2.5 py-2 text-[15px] outline-none"
                    />
                    <button
                        type="submit"
                        className="bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-4 rounded-lg transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        Add
                    </button>
                </form>

                {toBuy.length === 0 ? nothingToBuy : (
                    <div className="flex flex-col">
                        {toBuy.map(item => (
                            <div key={item.id} className="flex items-center gap-3 px-1 py-[11px] border-b border-stone-100">
                                <button
                                    onClick={() => handleToggle(item.id)}
                                    aria-label={`Tick off ${item.name}`}
                                    className="w-5 h-5 flex-none rounded-md border-[1.5px] border-stone-300 bg-white transition-colors hover:border-orange-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                />
                                <span className="text-[15px] capitalize">{item.name}</span>
                                <span className="text-[13px] text-stone-500">{quantity(item)}</span>
                                <div className="flex-1" />
                                <span className="text-xs text-stone-400">
                                    {item.source === 'meal_plan' ? 'meal plan' : 'manual'}
                                </span>
                                <button
                                    onClick={() => handleDelete(item.id)}
                                    aria-label={`Remove ${item.name}`}
                                    className="text-stone-300 hover:text-red-600 text-lg leading-none px-1 transition-colors"
                                >
                                    ×
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* In the basket */}
            <div className="flex flex-col gap-3 bg-stone-50 rounded-2xl px-5 py-[18px]">
                <div className="flex items-center justify-between">
                    <span className="text-base font-bold text-stone-600">
                        In the basket <span className="text-stone-400 font-medium">{basket.length}</span>
                    </span>
                    {basket.length > 0 && (
                        <button
                            onClick={handleDeleteChecked}
                            className="text-[13px] text-red-600 rounded transition-colors hover:text-red-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                        >
                            Clear
                        </button>
                    )}
                </div>

                {basket.length === 0 ? (
                    <span className="text-[13px] text-stone-400">Items you tick off land here.</span>
                ) : (
                    <div className="flex flex-col">
                        {basket.map(item => (
                            <div key={item.id} className="flex items-center gap-3 py-2">
                                <button
                                    onClick={() => handleToggle(item.id)}
                                    aria-label={`Put ${item.name} back on the list`}
                                    className="w-5 h-5 flex-none rounded-md border-[1.5px] border-orange-500 bg-orange-500 text-white text-xs font-bold flex items-center justify-center transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                                >
                                    ✓
                                </button>
                                <span className="text-sm text-stone-400 line-through capitalize">{item.name}</span>
                                <span className="text-[13px] text-stone-400">{quantity(item)}</span>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
        </>
    )
}

function ShoppingListSkeleton() {
    return (
        <div className="grid md:grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)] gap-7 items-start animate-pulse">
            <div className="flex flex-col gap-3">
                <div className="h-4 w-[90px] rounded-md bg-stone-100" />
                <div className="h-[50px] rounded-xl bg-stone-100" />
                {Array.from({ length: 6 }, (_, i) => (
                    <div key={i} className="flex items-center gap-3 py-[11px] border-b border-stone-100">
                        <div className="w-5 h-5 flex-none rounded-md bg-stone-100" />
                        <div className="h-[11px] w-[45%] rounded-md bg-stone-100" />
                    </div>
                ))}
            </div>

            <div className="flex flex-col gap-3 bg-stone-50 rounded-2xl px-5 py-[18px]">
                <div className="h-4 w-[120px] rounded-md bg-stone-100" />
                <div className="h-[11px] w-[70%] rounded-md bg-stone-100" />
            </div>
        </div>
    )
}

export default ShoppingList
