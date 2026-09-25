import { useState } from "react"
import toast from "react-hot-toast"
import { DIETARY_OPTIONS, INTOLERANCE_OPTIONS } from "../api/types"

export type Preferences = {
    diets: string[]
    intolerances: string[]
    excluded: string[]
}

export const emptyPreferences: Preferences = { diets: [], intolerances: [], excluded: [] }

type Props = {
    value: Preferences
    onChange: (next: Preferences) => void
    variant?: 'profile' | 'onboarding'
}

const chipBase = "rounded-full px-3.5 py-2 text-sm font-medium border transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"

const toggle = (list: string[], option: string) =>
    list.includes(option) ? list.filter(o => o !== option) : [...list, option]

function PreferencesForm({ value, onChange, variant = 'profile' }: Props) {
    const [ingredientInput, setIngredientInput] = useState('')

    const onboarding = variant === 'onboarding'

    const addIngredient = () => {
        const trimmed = ingredientInput.trim().toLowerCase()
        if (!trimmed) return
        if (value.excluded.includes(trimmed)) {
            toast.error('Already added')
            return
        }
        onChange({ ...value, excluded: [...value.excluded, trimmed] })
        setIngredientInput('')
    }

    const removeIngredient = (ingredient: string) => {
        onChange({ ...value, excluded: value.excluded.filter(i => i !== ingredient) })
    }

    return (
        <div className={`flex flex-col ${onboarding ? 'gap-6' : 'gap-7'}`}>

            <div className={`flex flex-col ${onboarding ? 'gap-2.5' : 'gap-3'}`}>
                <span className="text-sm font-semibold">Diet</span>
                <div className="flex flex-wrap gap-2">
                    {DIETARY_OPTIONS.map(option => {
                        const on = value.diets.includes(option)
                        return (
                            <button
                                key={option}
                                type="button"
                                onClick={() => onChange({ ...value, diets: toggle(value.diets, option) })}
                                className={`${chipBase} ${on
                                    ? 'bg-orange-500 border-orange-500 text-white'
                                    : 'bg-white border-stone-200 text-stone-600 hover:border-orange-300'
                                }`}
                            >
                                {option}
                            </button>
                        )
                    })}
                </div>
            </div>

            <div className={`flex flex-col ${onboarding ? 'gap-2.5' : 'gap-3'}`}>
                <span className="text-sm font-semibold">Intolerances</span>
                <div className="flex flex-wrap gap-2">
                    {INTOLERANCE_OPTIONS.map(option => {
                        const on = value.intolerances.includes(option)
                        return (
                            <button
                                key={option}
                                type="button"
                                onClick={() => onChange({ ...value, intolerances: toggle(value.intolerances, option) })}
                                className={`${chipBase} ${on
                                    ? 'bg-red-400 border-red-400 text-white'
                                    : 'bg-white border-stone-200 text-stone-600 hover:border-red-300'
                                }`}
                            >
                                {option}
                            </button>
                        )
                    })}
                </div>
            </div>

            <div className={`flex flex-col ${onboarding ? 'gap-2.5' : 'gap-3'}`}>
                <div className="flex flex-col gap-0.5">
                    <span className="text-sm font-semibold">Exclude ingredients</span>
                    {!onboarding && (
                        <span className="text-[13px] text-stone-500">
                            Allergies we missed, or anything you just don't want.
                        </span>
                    )}
                </div>

                <div className={`flex gap-1.5 border border-stone-200 rounded-xl p-[5px] ${onboarding ? '' : 'max-w-[420px]'}`}>
                    <input
                        type="text"
                        placeholder="e.g. mushrooms"
                        value={ingredientInput}
                        onChange={e => setIngredientInput(e.target.value)}
                        onKeyDown={e => {
                            if (e.key === 'Enter') {
                                e.preventDefault()
                                addIngredient()
                            }
                        }}
                        className="flex-1 min-w-0 border-none px-2.5 py-2 text-[15px] outline-none"
                    />
                    <button
                        type="button"
                        onClick={addIngredient}
                        className="bg-orange-100 hover:bg-orange-200 text-orange-700 text-sm font-semibold px-4 rounded-lg transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-orange-300"
                    >
                        Add
                    </button>
                </div>

                {value.excluded.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                        {value.excluded.map(ingredient => (
                            <span
                                key={ingredient}
                                className="flex items-center gap-1.5 bg-stone-100 text-stone-700 rounded-full pl-3.5 pr-2 py-1.5 text-sm"
                            >
                                {ingredient}
                                <button
                                    type="button"
                                    onClick={() => removeIngredient(ingredient)}
                                    aria-label={`Remove ${ingredient}`}
                                    className="text-stone-400 hover:text-red-600 text-base leading-none px-1 transition-colors"
                                >
                                    ×
                                </button>
                            </span>
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
}

export default PreferencesForm
