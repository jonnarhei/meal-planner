import client from "./client";
import type { MealPlan, MealPlanRecipe } from "./types";

export async function getCurrentMealPlan(): Promise<MealPlan> {
    const response = await client.get('/meal-plans/current') 
    return response.data
}

export async function changeRecipeForDay(day:number): Promise<MealPlanRecipe> {
    const response = await client.patch('/meal-plans/current/recipe', { day })
    return response.data
}

export async function regenerateMealPlan(): Promise<MealPlan> {
    const response = await client.post('/meal-plans/current/regenerate')
    return response.data
}