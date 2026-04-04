import client from "./client";
import type { ShoppingListItem } from "./types";

export async function getShoppingList(): Promise<ShoppingListItem[]> {
    const response = await client.get('/shopping-list')
    return response.data
}

export async function addItems(items :{ name: string, amount: number, unit: string }[]) {
    await client.post('/shopping-list/items', { items })
}

export async function addFromMealPlan(): Promise<void> {
    await client.post('/shopping-list/from-meal-plan')
}

export async function toggleChecked(id: number): Promise<void> {
    await client.patch(`/shopping-list/items/${id}`)
}

export async function deleteItem(id: number): Promise<void> {
    await client.delete(`/shopping-list/items/${id}`)
}

export async function deleteChecked(): Promise<void> {
    await client.delete('/shopping-list/checked')
}
