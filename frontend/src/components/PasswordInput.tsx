import { useState } from "react"

type Props = {
    value: string
    onChange: (value: string) => void
    placeHolder?: string
}

function PasswordInput({ value, onChange, placeHolder = 'Password' }: Props) {
    const [visible, setVisible] = useState(false)

    return (
        <div className="relative">
            <input
                type={visible ? 'text' : 'password'}
                placeholder={placeHolder}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className="w-full border border-gray-200 rounded-lg px-4 py-3 pr-16 text-sm focus:outline-none focus:ring-2 focus:ring-orange-300"
            />
            <button
                type="button"
                onClick={() => setVisible(!visible)}
                aria-label={visible ? `Hide ${placeHolder}` : `Show ${placeHolder}`}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-gray-500 hover:text-orange-500"
            >
                {visible ? 'Hide' : 'Show'}
            </button>
        </div>
    )
}

export default PasswordInput