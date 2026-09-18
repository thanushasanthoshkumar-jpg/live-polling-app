import React from 'react'

const COLORS = [
  'bg-indigo-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-sky-500',
  'bg-violet-500',
  'bg-teal-500',
  'bg-orange-500',
  'bg-pink-500',
  'bg-lime-500',
]

export default function PollResults({ options, counts }) {
  const total = options.reduce((sum, opt) => sum + (counts[opt.id] || 0), 0)

  return (
    <div className="space-y-3">
      {options.map((opt, i) => {
        const count = counts[opt.id] || 0
        const pct = total > 0 ? Math.round((count / total) * 100) : 0
        return (
          <div key={opt.id}>
            <div className="flex justify-between text-sm mb-1">
              <span className="text-gray-800 font-medium">{opt.text}</span>
              <span className="text-gray-500">
                {count} vote{count === 1 ? '' : 's'} ({pct}%)
              </span>
            </div>
            <div className="w-full bg-gray-100 rounded-full h-3 overflow-hidden">
              <div
                className={`h-3 rounded-full transition-all duration-500 ease-out ${COLORS[i % COLORS.length]}`}
                style={{ width: `${pct}%` }}
              />
            </div>
          </div>
        )
      })}
      <p className="text-xs text-gray-400 pt-1">{total} total vote{total === 1 ? '' : 's'}</p>
    </div>
  )
}
