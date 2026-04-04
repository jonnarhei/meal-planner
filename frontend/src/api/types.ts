export interface MealPlanRecipe {
    id: number
    meal_plan_id: number
    recipe_id: number
    recipe_title: string
    image: string
    source_url: string
    day: number
}

export interface MealPlan {
    id: number
    user_id: number
    start_date: string
    end_date: string
    created_at: string
    recipes: MealPlanRecipe[]
}

export interface User {
    id: number
    email: string
    dietary_preferences: string[]
}

export const DIETARY_OPTIONS = [
    'vegetarian',
    'vegan',
    'gluten free',
    'dairy free',
    'ketogenic',
    'paleo',
] as const

export interface ShoppingListItem {
    id: number
    user_id: number
    name: string
    amount: number
    unit: string
    checked: boolean
    source: string
    created_at: string
}